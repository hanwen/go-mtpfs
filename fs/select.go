package fs

import (
	"log"
	"regexp"
	
	"github.com/hanwen/go-mtpfs/mtp"
)

func SelectStorages(dev *mtp.Device, pat string) ([]uint32, error) {
	sids := mtp.Uint32Array{}
	err := dev.GetStorageIDs(&sids)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}

	filtered := []uint32{}
	for _, id := range sids.Values {
		var s mtp.StorageInfo
		err := dev.GetStorageInfo(id, &s)
		if err != nil {
			return nil, err
		}
		
		if !s.IsHierarchical() {
			log.Printf("skipping non hierarchical storage %q", s.StorageDescription)
			continue
		}

		if re.FindStringIndex(s.StorageDescription) == nil {
			log.Printf("filtering out storage %q", s.StorageDescription)
			continue
		}
		filtered = append(filtered, id)
	}

	return filtered, nil
}
