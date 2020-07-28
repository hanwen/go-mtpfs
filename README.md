## Introduction

![Diagram](./img/very_simple_diagram.jpg)

**日本語は[こちら](./README.ja.md)**

mtplvcap is a multi-platform (Windows/Mac/Linux) software that relays the Live View of Nikon cameras via WebSocket. 

mtplvcap + OBS turn your cameras into web cameras without HDMI capture device. Enjoy video chatting on Google Hangouts/Meet/Zoom etc. with your favorite cameras!


## Environments

### Cameras (verified)

 - Nikon D5300
 - Nikon D3300 (thanks [@unasuke](https://github.com/unasuke))
 - Nikon D7000
 - I welcome feedback! Please tell me if your camera works or not.


### Cameras (not working; investigating)

 - [Nikon D7200](https://github.com/puhitaku/mtplvcap/issues/2)
 - [Nikon 1 J5](https://github.com/puhitaku/mtplvcap/issues/4)


### Verified OSes and other software

|OS|OBS|Virtual camera|Browser|
|:-|:-|:-|:-|
|Windows 10 version 1909, Build 18363.900, amd64|25.0.8|OBS Virtualcam 2.0.5|Microsoft Edge 84.0.522.44 (Chromium Edge)|
|macOS 10.15.5 Catalina, amd64|25.0.8|obs-mac-virtualcam 3ca8f62 v1.2.0|Google Chrome 84.0.4147.89, Microsoft Edge 84.0.522.44|
|Debian GNU/Linux 10 Buster, amd64|25.0.7-442-ge3942061|obs-v4l2sink 1ec3c8a|Mozilla Firefox 68.10.0esr|

Note: Browsers listed here are used to check if virtual cams work well. Any other browsers should work.


## How to install

Notice: Snippets described here should run as-is and copy-and-pastable.


### Windows

**Important: For Windows, you have to replace a pre-installed MTP driver with a libusb driver.
Your PC will no longer recognize the camera as an MTP device unless you re-install it manually.
Continue with care.**


#### 1. Replace MTP driver

1. Connect your Nikon camera to the PC
1. Download Zadig from [here](https://zadig.akeo.ie/) and launch it
1. Tick `List All Devices`

    <img alt="Tick List All Devices" src="./img/zadig_1.png" width="400px">

1. Make sure that your camera is in the list upper in the window and choose it

    <img alt="Choose the camera in the list" src="./img/zadig_2.png" width="400px">

    (This screenshot was taken after the libusb driver is installed and will differ from what you see)

1. Choose `libusb-win32 (vX.X.X.X)` in the input box in the middle of the window
    - Please keep in mind that `WinUSB` does NOT work. Be careful not to choose it.

1. Click `Replace Driver` button and wait it finishes the installation
    - Optionally, open the Device Manager and make sure it's installed

    <img alt="Device Manager after the installation" src="./img/devmgmt.png" width="400px">


#### 2a. Use a pre-built executable

1. Download the release from [here](https://github.com/puhitaku/mtplvcap/releases) (mtplvcap_windows_amd64.zip).
1. Extract the ZIP
1. Double-click `mtplvcap.exe`
    - Make sure your camera opens up its shutter


#### 2b. Build it yourself in MSYS2

1. Download and install MSYS2 from [here](https://www.msys2.org/)
1. Launch "MSYS2 MSYS" in the Start Menu
1. Install dependencies

    - Some people reported that pacman fails to install a package group `mingw-w64-x86_64-toolchain` and other packages at once. Installing them separately should succeed.

    ```sh
    pacman -Sy
    pacman -S mingw-w64-x86_64-toolchain
    pacman -S \
        mingw-w64-x86_64-libusb \
        mingw-w64-x86_64-go \
        mingw-w64-x86_64-pkg-config \
        git
    ```

1. Add PATHs

    ```sh
    echo 'PATH=$PATH:/mingw64/bin:/mingw64/lib/go/bin' >> ~/.bashrc
    source ~/.bashrc
    ```

1. Clone this repo

    ```sh
    git clone https://github.com/puhitaku/mtplvcap.git
    ```

1. `cd`, build, and launch it

    ```sh
    cd mtplvcap
    CGO_CFLAGS='-Wno-deprecated-declarations' go build .
    ./mtplvcap.exe -debug server
    ```
   
    - Make sure your camera opens up its shutter
    - Perhaps you have to specify GOROOT like `GOROOT=/mingw64/lib/go go build .`

1. Done!
    - The binary can be moved and redistributed easily
    - Copy `libusb-1.0.dll` from `C:\msys64\mingw64\bin\libusb-1.0.dll` and place the copy alongside `mtplvcap.exe` to launch it directly from Explorer


### macOS

#### 1. Install dependencies

1. [Install Homebrew](https://brew.sh/) 

1. Install libusb

    ```sh
    brew install libusb
    ```


#### 2a. Use a pre-built executable

1. Download the release from [here](https://github.com/puhitaku/mtplvcap/releases) (mtplvcap_macos_amd64.zip).
1. Extract the ZIP and launch it

    ```sh
    unzip mtplvcap_macos_amd64.zip
    ./macos/mtplvcap
    ```
    - macOS will warn you that the binary is not verified

1. Allow non-verified binary to launch

    - Dismiss the dialog (Do not move it to the trash!)

    <img alt="macOS warning dialog" src="./img/macos_warning.png" width="400px">

    - Open "System Preferences" -> "Security & Privacy" -> Click "Allow Anyway"

    <img alt="System Preference" src="./img/macos_warning2.png" width="400px">

    - Launch `mtplvcap` once more and click "Open"

    <img alt="macOS warning dialog 2" src="./img/macos_warning3.png" width="400px">

    - Make sure your camera opens up its shutter


#### 2b. Build it yourself

1. Install dependencies

    ```sh
    brew install golang git
    ```

1. Install XCode Command Line Tools with `xcode-select --install`

1. Clone this repo

    ```sh
    git clone https://github.com/puhitaku/mtplvcap.git
    ```

1. `cd`, build, and launch it

    ```sh
    cd mtplvcap
    CGO_CFLAGS='-Wno-deprecated-declarations' go build .
    ./mtplvcap -debug server
    ```

    - If macOS blocks the downloaded executable and show alerts, please follow [this procedure](https://github.com/hashicorp/terraform/issues/23033#issuecomment-542302933) to proceed
    - Make sure your camera opens up its shutter

1. Done!


### Linux

I strongly recommend you to compile by yourself for Linux distributions as Linux environments vary widely.


#### 1. Install dependencies

1. Install libusb

    ```sh
    # For Debian/Ubuntu:
    sudo apt install libusb-1.0.0
    ```

    ```sh
    # For Arch:
    pacman --sync libusb
    ```


#### 2a. Use a pre-built executable

1. Download the release from [here](https://github.com/puhitaku/mtplvcap/releases) (mtplvcap_linux_amd64.zip).

1. Extract the ZIP and launch it

    ```sh
    unzip mtplvcap_linux_amd64.zip
    ./linux/mtplvcap
    ```

    - Make sure your camera opens up its shutter


#### 2b. Build it yourself

1. Install dependencies

    ```sh
    # For Debian/Ubuntu:
    sudo apt install golang-go libusb-1.0.0-dev
    ```

    ```sh
    # For Arch:
    pacman --sync go libusb
    ```

1. `cd`, build, and launch it

    ```sh
    cd mtplvcap
    CGO_CFLAGS='-Wno-deprecated-declarations' go build .
    ./mtplvcap -debug server
    ```

    - Make sure your camera opens up its shutter


### Usage

```sh
$ ./mtplvcap -help
Usage of ./mtplvcap:
  -backend-go
        force gousb as libusb wrapper (not recommended)
  -debug string
        comma-separated list of debugging options: usb, data, mtp, server
  -host string
        hostname: default = localhost, specify 0.0.0.0 for public access (default "localhost")
  -port int
        port: default = 42839 (default 42839)
  -product-id string
        PID of the camera to search (in hex), default=0x0 (all) (default "0x0")
  -server-only
        serve frontend without opening a DSLR (for devevelopment)
  -vendor-id string
        VID of the camera to search (in hex), default=0x0 (all) (default "0x0")
```


#### Watch incoming frames

 - `http://localhost:42839/view` will show the captured frames


#### Control your camera on your browser

 - `http://localhost:42839` is a controller to control your camera
 - "Auto Focus" section controls periodic/manual AF
 - "Rate Limit" section limits/un-limits the frame rate to decrease overall CPU usage
 - "Information" section shows the dimension of captured images etc.


#### Connect with Zoom, Google Meet, Google Hangouts, etc.

1. Install mtplvcap and check if it works

1. Install OBS (Open Broadcaster Software) from [here](https://obsproject.com/)

1. Install OBS virtual camera (it varies for each OS)

1. Open OBS preference and "Video" tab

1. Adjust resolutions to fit with LV frame dimension
    - Launch mtplvcap and open `localhost:42839` to get the actual resolution

    <img alt="Controller view" src="./img/obs_1.png" width="400px">
    <img alt="Adjust resolution" src="./img/obs_2.png" width="400px">

1. Add a "Browser" source

    <img alt="Add Browser source" src="./img/obs_3.png" width="400px">

1. Set "Width" and "Height" to the same value of "Video" preference

1. Set `http://localhost:42839/view` as the URL

    <img alt="Set URL" src="./img/obs_4.png" width="400px">

1. Enable the virtual camera and configure chat apps

1. BOOM!

    <img alt="Hi!" src="./img/obs_5.png" width="400px">
    <img alt="Zoom!" src="./img/obs_6.png" width="400px">


### Caveats

 - This software is in alpha stage.


### Known Issues

 - Stops Live View automatically
    - The timeout can be elongated in camera menu.
    - Some cameras support disabling auto-off timer, while some does not.
        - For D5300, "Custom Setting Menu" -> "c Timers/AE lock" -> "Auto off timers" -> "Custom" -> "Live View" -> "30 min"
    - As a work-around, mtplvcap watches the shutter and opens it automatically when it's closed.
 - Windows: on MinTTY, the process gets killed without graceful shut-down when you press Ctrl-C.
    - It will result in a fail of the next launch and might require you to re-plug the camera in.
    - It's a known behavior and is not a bug of mtplvcap.
    - Please install winpty with pacman and run via it `winpty ./mtplvcap`
    - Running mtplvcap directly from Explorer with double-click runs without this problem. Please stop it with Ctrl-C, not by pressing the close button.


### Feedback

 - Posting issues and PRs is welcome. Follow [CONTRIBUTING.md](./CONTRIBUTING.md) for contribution.
 - Only few cameras are verified. Please tell me if mtplvcap works (or not) with your camera.


### Credit

This program is based on [github.com/hanwen/go-mtpfs](https://github.com/hanwen/go-mtpfs).
Special thanks to Han-Wen-san for a robust and mature MTP implementation.

And also I referenced [github.com/dukus/digiCamControl](https://github.com/dukus/digiCamControl) for parsing MTP payloads.
If not for this, I couldn't have implemented mtplvcap.

### License

[LICENSE document](./LICENSE)
