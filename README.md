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

----

## All endpoints

###  activation

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /{udid}/activated | [get udid activated](#get-udid-activated) | Check activation status |
| POST | /{udid}/activate/enable | [post udid activate enable](#post-udid-activate-enable) | Enable activation |
  


###  apps

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /{udid}/apps/list | [get udid apps list](#get-udid-apps-list) | List applications |
| POST | /{udid}/apps/install | [post udid apps install](#post-udid-apps-install) | Install application |
| POST | /{udid}/apps/kill | [post udid apps kill](#post-udid-apps-kill) | Kill application |
| POST | /{udid}/apps/run | [post udid apps run](#post-udid-apps-run) | Run application |
  


###  developer

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /{udid}/devmode | [get udid devmode](#get-udid-devmode) | Check developer mode status |
| GET | /{udid}/image | [get udid image](#get-udid-image) | Check developer disk image status |
| POST | /{udid}/devmode/enable | [post udid devmode enable](#post-udid-devmode-enable) | Enable developer mode |
| POST | /{udid}/image/enable | [post udid image enable](#post-udid-image-enable) | Mount developer disk image |
  


###  device

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /devices | [get devices](#get-devices) | List devices |
| GET | /{udid}/processes | [get udid processes](#get-udid-processes) | List processes |
| POST | /{udid}/erase | [post udid erase](#post-udid-erase) | Erase device |
| POST | /{udid}/reboot | [post udid reboot](#post-udid-reboot) | Reboot device |
  


###  pairing

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /{udid}/paired | [get udid paired](#get-udid-paired) | Check pairing status |
| POST | /{udid}/pair/enable | [post udid pair enable](#post-udid-pair-enable) | Enable pairing |
  


###  profiles

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /{udid}/profiles/list | [get udid profiles list](#get-udid-profiles-list) | List profiles |
| POST | /{udid}/profiles/add | [post udid profiles add](#post-udid-profiles-add) | Add profile |
  


###  supervision

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /{udid}/supervised | [get udid supervised](#get-udid-supervised) | Check supervision status |
| POST | /{udid}/supervise/enable | [post udid supervise enable](#post-udid-supervise-enable) | Enable supervision |
  


###  wda

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /{udid}/wda/kill | [post udid wda kill](#post-udid-wda-kill) | Kill WebDriverAgent |
| POST | /{udid}/wda/run | [post udid wda run](#post-udid-wda-run) | Run WebDriverAgent |
  


## Paths

### <span id="get-devices"></span> List devices (*GetDevices*)

```
GET /devices
```

Returns a list of all connected iOS devices

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-devices-200) | OK | OK |  | [schema](#get-devices-200-schema) |

#### Responses


##### <span id="get-devices-200"></span> 200 - OK
Status: OK

###### <span id="get-devices-200-schema"></span> Schema
   
  

[MainDevicesResponse](#main-devices-response)

### <span id="get-udid-activated"></span> Check activation status (*GetUdidActivated*)

```
GET /{udid}/activated
```

Returns whether the device is activated

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-udid-activated-200) | OK | OK |  | [schema](#get-udid-activated-200-schema) |

#### Responses


##### <span id="get-udid-activated-200"></span> 200 - OK
Status: OK

###### <span id="get-udid-activated-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="get-udid-apps-list"></span> List applications (*GetUdidAppsList*)

```
GET /{udid}/apps/list
```

Returns a list of applications installed on the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-udid-apps-list-200) | OK | OK |  | [schema](#get-udid-apps-list-200-schema) |

#### Responses


##### <span id="get-udid-apps-list-200"></span> 200 - OK
Status: OK

###### <span id="get-udid-apps-list-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="get-udid-devmode"></span> Check developer mode status (*GetUdidDevmode*)

```
GET /{udid}/devmode
```

Returns whether developer mode is enabled on the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-udid-devmode-200) | OK | OK |  | [schema](#get-udid-devmode-200-schema) |

#### Responses


##### <span id="get-udid-devmode-200"></span> 200 - OK
Status: OK

###### <span id="get-udid-devmode-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="get-udid-image"></span> Check developer disk image status (*GetUdidImage*)

```
GET /{udid}/image
```

Returns whether the developer disk image is mounted

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-udid-image-200) | OK | OK |  | [schema](#get-udid-image-200-schema) |

#### Responses


##### <span id="get-udid-image-200"></span> 200 - OK
Status: OK

###### <span id="get-udid-image-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="get-udid-paired"></span> Check pairing status (*GetUdidPaired*)

```
GET /{udid}/paired
```

Returns whether the device is paired

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-udid-paired-200) | OK | OK |  | [schema](#get-udid-paired-200-schema) |

#### Responses


##### <span id="get-udid-paired-200"></span> 200 - OK
Status: OK

###### <span id="get-udid-paired-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="get-udid-processes"></span> List processes (*GetUdidProcesses*)

```
GET /{udid}/processes
```

Returns a list of running processes on the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-udid-processes-200) | OK | OK |  | [schema](#get-udid-processes-200-schema) |

#### Responses


##### <span id="get-udid-processes-200"></span> 200 - OK
Status: OK

###### <span id="get-udid-processes-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="get-udid-profiles-list"></span> List profiles (*GetUdidProfilesList*)

```
GET /{udid}/profiles/list
```

Returns a list of configuration profiles installed on the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-udid-profiles-list-200) | OK | OK |  | [schema](#get-udid-profiles-list-200-schema) |

#### Responses


##### <span id="get-udid-profiles-list-200"></span> 200 - OK
Status: OK

###### <span id="get-udid-profiles-list-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="get-udid-supervised"></span> Check supervision status (*GetUdidSupervised*)

```
GET /{udid}/supervised
```

Returns whether the device is supervised

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-udid-supervised-200) | OK | OK |  | [schema](#get-udid-supervised-200-schema) |

#### Responses


##### <span id="get-udid-supervised-200"></span> 200 - OK
Status: OK

###### <span id="get-udid-supervised-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-activate-enable"></span> Enable activation (*PostUdidActivateEnable*)

```
POST /{udid}/activate/enable
```

Activates the specified iOS device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-activate-enable-200) | OK | OK |  | [schema](#post-udid-activate-enable-200-schema) |

#### Responses


##### <span id="post-udid-activate-enable-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-activate-enable-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-apps-install"></span> Install application (*PostUdidAppsInstall*)

```
POST /{udid}/apps/install
```

Installs an application from a URL on the device

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |
| request | `body` | [MainAppInstallRequest](#main-app-install-request) | `models.MainAppInstallRequest` | | ✓ | | Application IPA URL |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-apps-install-200) | OK | OK |  | [schema](#post-udid-apps-install-200-schema) |
| [400](#post-udid-apps-install-400) | Bad Request | invalid JSON |  | [schema](#post-udid-apps-install-400-schema) |

#### Responses


##### <span id="post-udid-apps-install-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-apps-install-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

##### <span id="post-udid-apps-install-400"></span> 400 - invalid JSON
Status: Bad Request

###### <span id="post-udid-apps-install-400-schema"></span> Schema
   
  



### <span id="post-udid-apps-kill"></span> Kill application (*PostUdidAppsKill*)

```
POST /{udid}/apps/kill
```

Terminates a running application by process ID

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |
| pid | `formData` | string | `string` |  | ✓ |  | Process ID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-apps-kill-200) | OK | OK |  | [schema](#post-udid-apps-kill-200-schema) |

#### Responses


##### <span id="post-udid-apps-kill-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-apps-kill-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-apps-run"></span> Run application (*PostUdidAppsRun*)

```
POST /{udid}/apps/run
```

Launches an application on the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |
| bundleid | `formData` | string | `string` |  | ✓ |  | Application bundle identifier |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-apps-run-200) | OK | OK |  | [schema](#post-udid-apps-run-200-schema) |

#### Responses


##### <span id="post-udid-apps-run-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-apps-run-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-devmode-enable"></span> Enable developer mode (*PostUdidDevmodeEnable*)

```
POST /{udid}/devmode/enable
```

Enables developer mode on the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-devmode-enable-200) | OK | OK |  | [schema](#post-udid-devmode-enable-200-schema) |

#### Responses


##### <span id="post-udid-devmode-enable-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-devmode-enable-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-erase"></span> Erase device (*PostUdidErase*)

```
POST /{udid}/erase
```

Erases all content and settings from the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-erase-200) | OK | OK |  | [schema](#post-udid-erase-200-schema) |

#### Responses


##### <span id="post-udid-erase-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-erase-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-image-enable"></span> Mount developer disk image (*PostUdidImageEnable*)

```
POST /{udid}/image/enable
```

Mounts the developer disk image on the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-image-enable-200) | OK | OK |  | [schema](#post-udid-image-enable-200-schema) |

#### Responses


##### <span id="post-udid-image-enable-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-image-enable-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-pair-enable"></span> Enable pairing (*PostUdidPairEnable*)

```
POST /{udid}/pair/enable
```

Pairs the device using the provided certificate

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-pair-enable-200) | OK | OK |  | [schema](#post-udid-pair-enable-200-schema) |

#### Responses


##### <span id="post-udid-pair-enable-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-pair-enable-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-profiles-add"></span> Add profile (*PostUdidProfilesAdd*)

```
POST /{udid}/profiles/add
```

Installs a configuration profile on the device

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |
| profile | `body` | [MainProfileAddRequest](#main-profile-add-request) | `models.MainProfileAddRequest` | | ✓ | | Base64 encoded profile |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-profiles-add-200) | OK | OK |  | [schema](#post-udid-profiles-add-200-schema) |
| [400](#post-udid-profiles-add-400) | Bad Request | invalid JSON |  | [schema](#post-udid-profiles-add-400-schema) |

#### Responses


##### <span id="post-udid-profiles-add-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-profiles-add-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

##### <span id="post-udid-profiles-add-400"></span> 400 - invalid JSON
Status: Bad Request

###### <span id="post-udid-profiles-add-400-schema"></span> Schema
   
  



### <span id="post-udid-reboot"></span> Reboot device (*PostUdidReboot*)

```
POST /{udid}/reboot
```

Reboots the specified iOS device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-reboot-200) | OK | OK |  | [schema](#post-udid-reboot-200-schema) |

#### Responses


##### <span id="post-udid-reboot-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-reboot-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-supervise-enable"></span> Enable supervision (*PostUdidSuperviseEnable*)

```
POST /{udid}/supervise/enable
```

Prepares and enables supervision on the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-supervise-enable-200) | OK | OK |  | [schema](#post-udid-supervise-enable-200-schema) |

#### Responses


##### <span id="post-udid-supervise-enable-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-supervise-enable-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-wda-kill"></span> Kill WebDriverAgent (*PostUdidWdaKill*)

```
POST /{udid}/wda/kill
```

Stops WebDriverAgent on the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-wda-kill-200) | OK | OK |  | [schema](#post-udid-wda-kill-200-schema) |

#### Responses


##### <span id="post-udid-wda-kill-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-wda-kill-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

### <span id="post-udid-wda-run"></span> Run WebDriverAgent (*PostUdidWdaRun*)

```
POST /{udid}/wda/run
```

Starts WebDriverAgent on the device

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| udid | `path` | string | `string` |  | ✓ |  | Device UDID |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#post-udid-wda-run-200) | OK | OK |  | [schema](#post-udid-wda-run-200-schema) |

#### Responses


##### <span id="post-udid-wda-run-200"></span> 200 - OK
Status: OK

###### <span id="post-udid-wda-run-200-schema"></span> Schema
   
  

[MainGenericResponse](#main-generic-response)

## Models

### <span id="main-app-install-request"></span> main.AppInstallRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| url | string| `string` |  | |  |  |



### <span id="main-device"></span> main.Device


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| ConnectionType | string| `string` |  | |  |  |
| ProductName | string| `string` |  | |  |  |
| ProductType | string| `string` |  | |  |  |
| ProductVersion | string| `string` |  | |  |  |
| Udid | string| `string` |  | |  |  |



### <span id="main-devices-response"></span> main.DevicesResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| devices | [][MainDevice](#main-device)| `[]*MainDevice` |  | |  |  |



### <span id="main-generic-response"></span> main.GenericResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| ok | boolean| `bool` |  | |  |  |



### <span id="main-profile-add-request"></span> main.ProfileAddRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| b64profile | string| `string` |  | |  |  |