# Architecture

The package is split into a background service and a tui application. The
service is responsible for proxying communication between the user and other
users. The reason for this split is to enable the receiving of messages, even
when the tui application isn't running. More critically, it makes the user
discoverable and reachable before they open the tui. The tui can be used to
enable or disable the service.
