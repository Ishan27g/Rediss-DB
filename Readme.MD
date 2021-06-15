### All replicas available via Web socket or commandline
```text
- ws://localhost:9001/ws
- ws://localhost:9003/ws
- ws://localhost:9003/ws
```

### Commands
#### Add : >Key=Value 
```text
>ID=123 {"ID" : "123"}
```
#### Get : <Key
```text
<ID // {"ID" : "123"}
```
#### Delete : -KEY
```text
-ID // {"Ok"}
```