# Provider Discovery
```
NewControlPanel()
       ↓
go telemService.FindProvider() - has a callback for onFind
       |
       |iterates over the known providers
       |
     onFind
       ↓
telemService.SwitchProvider() - activates the found provider
       |
       |-→ Create a background job while the stream hasn't initiated to listened
       |   for data, otherwise the connection might die before we start streaming
       |                                 ↓
       |                     go telemService.ProviderMonitor()
       ↓                     |                              |           
  /-→waits--\    if the provider stops we        this healthcheck is stopped 
  \_________/    we clear the provider and       once the stream starts
                 go back to the                   
                             ↓
                       FindProvider()
```
The provider discovery routine starts on `tui/tui.go`.
`FindProvider` is called here and it starts a background job.
On successful discovery the background job calls the `SwitchProvider` method
and dies.

## Provider stalls
A provider stalls once there is no new data.
After stalling
