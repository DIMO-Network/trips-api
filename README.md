# trips-api

```mermaid
flowchart LR
  topic.device.trip.event --> trips-api
  topic.event --> trips-api
```

Trip opening:
```
{
  …,
  data: {
    id: "2XUPU7gd9TWnkwXLH6MuzxTMnCD",
    deviceId: "2XUPU7gd9TWnkwXLH6MuzxTMnCD",
    completed: false
    start: {
      point: {
        latitude: 42.33203898012946,
        longitude: -83.87908201282218
      },
      time: "2023-10-30T09:22:21Z"
    }
  }
}
```
