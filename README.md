tinyios is cross platform software to talk to ios devices over usb or wifi. It mimics a small set of features that can be done via xcode, developer tools and system settings on a mac.

It exposes itself as a HTTP server so it can be accessed from anywhere. It is also stateless and rootless which makes it easy to run as an ephemeral container.

tinyios depends on usbmuxd installed on the host machine to manage device pairing and device communication through usb or wifi. You will also need to make usbmuxd on the host available inside the container.

Permissions and privileges:
Stateless: No tunnel manager or file system access
Rootless: No root access
