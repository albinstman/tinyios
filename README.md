# tinyios

[![Go Report Card](https://goreportcard.com/badge/github.com/albinstman/tinyios)](https://goreportcard.com/report/github.com/albinstman/tinyios) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/albinstman/tinyios?sort=semver)

----

tinyios is cross platform software to talk to ios devices over usb or wifi. It mimics a small set of features that can be done via xcode, developer tools and system settings on a mac.

It exposes itself as a HTTP server so it can be accessed from anywhere. It is also stateless and rootless which makes it easy to run as an ephemeral container.

tinyios depends on usbmuxd installed on the host machine to manage device pairing and device communication through usb or wifi. You will also need to make usbmuxd on the host available inside the container.

----
## Installation

### WSL

1. Forward device to WSL
2. Install dependancies

    To use tinyios on Linux or WSL you need to install usbmuxd to handle device pairing and device communication. You can run these commands to install or update existing usbmuxd to the latest version

    ```
    git clone https://github.com/libimobiledevice/usbmuxd.git
    cd usbmuxd
    ./autogen.sh
    make
    sudo make install
    ```
    You will also need socat to make usbmuxd available from the host to container
    ```
    sudo apt install socat
    sudo socat TCP-LISTEN:27015,reuseaddr,fork UNIX-CONNECT:/var/run/usbmuxd
    ```
3. Run latest container image on port 8080
    ```
    docker run —rm \
    -p 8080:80 \
    -e USBMUXD_SOCKET_ADDRESS=host.docker.internal:27015 \
    albinstman/tinyios
    ```

### Linux

1. Install dependancies

    To use tinyios on Linux or WSL you need to install usbmuxd to handle device pairing and device communication. You can run these commands to install or update existing usbmuxd to the latest version

    ```
    git clone https://github.com/libimobiledevice/usbmuxd.git
    cd usbmuxd
    ./autogen.sh
    make
    sudo make install
    ```
    You will also need socat to make usbmuxd available from the host to container
    ```
    sudo apt install socat
    sudo socat TCP-LISTEN:27015,reuseaddr,fork UNIX-CONNECT:/var/run/usbmuxd
    ```
2. Run latest container image on port 8080
    ```
    docker run —rm \
    -p 8080:80 \
    -e USBMUXD_SOCKET_ADDRESS=host.docker.internal:27015 \
    albinstman/tinyios
    ```

### Mac

1. Make usbmuxd available for container
    ```
    brew install socat
    socat TCP-LISTEN:27015,reuseaddr,fork UNIX-CONNECT:/var/run/usbmuxd
    ```
2. Run latest container image on port 8080
    ```
    docker run —rm \
    -p 8080:80 \
    -e USBMUXD_SOCKET_ADDRESS=host.docker.internal:27015 \
    albinstman/tinyios
    ```

----

## Purpose 
The main purpose of tinyios is to setup ios devices and then talk to them via appium webdriver commands.