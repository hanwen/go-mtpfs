# transform ptp.h into Go.
# $ python munge.py ~/vc/libmtp/src/ptp.h > mtp/const.go && gofmt -w mtp/const.go

import re
import sys

def toInt(v):
    v = v.lower()
    if v.startswith('0x'):
        return int(v[2:], 16)
    assert not v.startswith('0')
    return int(v)

expanation = {
    'DPC': 'device property code',
    'DPFF': 'device property form field',
    'DPGS': 'device property get/set',
    'DTC': 'data type code',
    'EC': 'event code',
    'EC': 'event code',
    'GOH': 'get object handles',
    'OC': 'operation code',
    'OFC': 'object format code',
    'OPC': 'object property code',
    'RC': 'return code',
    'ST': 'storage',
}

# prefix -> (name => val)
data = {}

for l in open(sys.argv[1]):
    m = re.match('^#define[ \t]+([A-Z0-9_a-z]+)[ \t]+((0x)?[0-9a-fA-F]+)$', l)
    if not m:
        continue

    name = m.group(1)
    val = m.group(2)

    if not name.startswith('PTP_'):
        continue

    name = name[4:]
    if '_' not in name :
        continue
        
    prefix, suffix = name.split('_', 1)
    if prefix not in data:
        data[prefix] = {}

    data[prefix][suffix] = val

del data['CANON']
vendors = set(data['VENDOR'].keys())
# sys.stderr.write('vendors %s' % vendors)

# brilliant.
vendors.add('EK')

print 'package mtp'
print '// DO NOT EDIT : generated automatically by munge.py'
for p, nv in sorted(data.items()):
    if p in expanation:
        print '\n// %s' % expanation[p]

    names = []
    seen = set()
    for v, n in sorted((v, n) for n, v in nv.items()):
        print 'const %s_%s = %s' % (p, n, v)
        iv = toInt(v)
        if iv in seen:
            continue
        seen.add(iv)
        if '_' in n:
            prefix, _ = n.split('_', 1)
            if prefix in vendors:
                #sys.stderr.write('vendor ext %s\n' % n)
                continue
        
        names.append('%s: "%s",\n' % (v, n))
        
    print '''var %s_names = map[int]string{%s}''' %(
        p,
        ''.join(names))

