# Automatic Reconnect

- RigClient tracks the current connection state using the conn field. If the field is nil, the state is "Disconnected", if the field is not nil, the state is connected.
- RigClient currently establishes the connection in NewRigClient. This needs to be moved into a separate Open method. RigClient is initially disconnected.
- The Open method establishes the connection. If this fails, it returns an error and sets the conn field to nil.
- RigClient has a Close method that closes the connection on purpose. It sets the conn field to nil.
- RigClient has an IsConnected method that returns the current connection state (true == connected, false == disconnected).
- When any request to rigctld is issued and RigClient is currently disconnected and the automatic reconnect is not activated, the corresponding RigClient method returns an error.
- When any request to rigctld is issued and RigClient is currently disconnected and the automatic reconnect is activated, RigClient tries to reconnect before sending the request to rigctld.
- RigClient has a Notify(any) function that appends the given object to an internal slice field named listeners of type any.
- The new interface type named RigConnectionListener has only one method named RigConnected(bool). The bool parameter indicates if the connection is established or not (true == connected, false == disconnected).
- When the connection state of RigClient changes, it iterates through the listener slice and checks if the current entry implements the RigConnectionListener interface. If so, the entry is casted to RigConnectionListener and its RigConnected method is called with the new connection state as parameter (true == connected, false == disconnected).
- The Open method has a boolean parameter to enable the automatic reconnection. If this parameter is set to false, and the connection is lost, the connection listener gets notified and RigClient stays disconnected. If this parameter is set to true, the connection listener gets notified and RigClient stays disconnected until the next request to rigctld is issued.
