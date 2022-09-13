import "./App.css";
import ReactMapboxGl, { Layer } from "react-mapbox-gl";
import React, { useEffect, useState } from "react";
import { json } from "stream/consumers";
import * as dotenv from "dotenv";

const MapBox = ReactMapboxGl({
  accessToken:
    "pk.eyJ1IjoiYmVua29jaGFub3dza2kiLCJhIjoiY2t6eDdlZzRnOGUyeTJvbXphdXdvZnJjZSJ9.S1WS1tkPKQnGt3A5Y72ZUA",
  attributionControl: false,
  doubleClickZoom: false,
  // maxZoom: 7,
  // minZoom: 2,
});
const styleUrl = "mapbox://styles/mapbox/dark-v10";

interface DeviceInfo {
  lat: number;
  lon: number;
  order: number;
  timestamp: string;
  speed: number;
  odometer: number;
  chargeRange: number;
}

interface DeviceTrips {
  trip_id: string;
  device_id: string;
  trip_start: string;
  trip_end: string;
}

interface DeviceMetaData {
  lat: number;
  lon: number;
  deviceID: string;
  zoom?: number;
}

interface Coordinates {
  lat: number;
  lon: number;
  zoom?: number;
}

function App() {
  const [deviceIDs, setDeviceIDs] = useState<string[]>([]);
  // const [deviceInfo, setDeviceInfo] = useState<DeviceInfo[]>([]);
  const [deviceTripInfo, setDeviceTripInfo] = useState<DeviceTrips[]>(
    {} as DeviceTrips[]
  );
  const [allDeviceTrips, setAllDeviceTrips] = useState<DeviceTrips[]>(
    {} as DeviceTrips[]
  );
  const [mapCenter, setMapCenter] = useState<Coordinates>({} as Coordinates);
  const [selectedDevice, setSelectedDevice] = useState("");
  const [mapkey, setMapKey] = useState(0);
  const [myMap, setMyMap] = useState<mapboxgl.Map>();

  const [tileserver, setTileserver] = useState(
    "http://localhost:7800/public.trips_odometer,public.points_odometer,public.trips_speed/{z}/{x}/{y}.pbf"
  );

  // const toggleLayer = (layerID: string) => {
  //   const visiblity = myMap?.getLayoutProperty(layerID, "visibility") as string;
  //   if (visiblity === "visible") {
  //     myMap?.setLayoutProperty(layerID, "visibility", "none");
  //   } else {
  //     myMap?.setLayoutProperty(layerID, "visibility", "visible");
  //   }
  // };

  useEffect(() => {
    if (selectedDevice !== "") {
      setTileserver(
        "http://localhost:7800/public.trips_odometer,public.points_odometer,public.trips_speed/{z}/{x}/{y}.pbf?device_key=" +
          selectedDevice
      );
      setMapKey(mapkey + 1);
    }
  }, [selectedDevice]);

  useEffect(() => {
    fetch("http://localhost:8000/devices/all")
      .then((r) => r.json())
      .then((data) => setDeviceIDs(data.sort() as []))
      .catch((e) => console.error(e));
  }, []);

  // useEffect(() => {
  //   if (selectedDevice !== "") {
  //     fetch(`http://localhost:8000/odometertrip/${selectedDevice}`)
  //       .then((r) => r.json())
  //       .then((data) => setDeviceInfo(data as []))
  //       .catch((e) => console.error(e));
  //   }
  // }, [selectedDevice]);

  useEffect(() => {
    if (selectedDevice !== "") {
      fetch(`http://localhost:8000/devices/${selectedDevice}/ongoing`, {
        method: "GET",
        headers: new Headers({
          Authorization: "Bearer ",
        }),
      })
        .then((r) => {
          return r.json();
        })
        .then((data) => setDeviceTripInfo(data))
        .catch((e) => console.error(e));
    }
  }, [selectedDevice]);

  useEffect(() => {
    if (selectedDevice !== "") {
      fetch(`http://localhost:8000/devices/${selectedDevice}/alltrips`, {
        method: "GET",
        headers: new Headers({
          Authorization:
            "Bearer " +
            "eyJhbGciOiJSUzI1NiIsImtpZCI6IjU1ZDhlMDU2MzBmMmI1OGFhMjQzNTVlOTM2OWE1OGQ3NzU3MjFmMzMifQ.eyJpc3MiOiJodHRwOi8vMTI3LjAuMC4xOjU1NTYvZGV4IiwicHJvdmlkZXJfaWQiOiJtb2NrIiwic3ViIjoiQ2cwd0xUTTROUzB5T0RBNE9TMHdFZ1J0YjJOciIsImF1ZCI6ImV4YW1wbGUtYXBwIiwiZXhwIjoxNjYyNzk4NTk1LCJpYXQiOjE2NjI3NTUzOTUsImF0X2hhc2giOiJ2OWhrVkwtLUR6cHRaZndJUXFuWkVRIiwiY19oYXNoIjoiR2VYNzN3Nk9PRC0tb1JVejcwcGlvUSIsImVtYWlsIjoia2lsZ29yZUBraWxnb3JlLnRyb3V0IiwiZW1haWxfdmVyaWZpZWQiOnRydWUsIm5hbWUiOiJLaWxnb3JlIFRyb3V0In0.syvb9LQdfYoTwRXmlfp8siokzu5pt6QTNgXWfxJV8AyvLNhp7flbMe0NXioPlOOtJv4m-rZdU6cx3wDqk0h3RGCJR-GHABJfjF7djiFNI14Hk_Ox2wceQxUYpiXkhGDLYANfBBy-atDTlwdHtar2Vz-zyBogvwcoL-tB9toXZCxbLulmzLfPDndTsHFOoOzPiLul1Xn_OGdX2rxASyXb99Ed0heGk3kJ1jicEqhDEA8cBGGvxq-7k04WlVLtXZTv-VhCbRud4cfnfef8-3eS8kkRYiyPnNLIVTq8-b0qeyRVXUkB_hdY_ewuN_ke9IyncL-nKWKS8DyO3Od74d8cVQ",
        }),
      })
        .then((r) => {
          return r.json();
        })
        .then((data) => setAllDeviceTrips(data))
        .catch((e) => console.error(e));
    }
  }, [selectedDevice]);

  const addHeatMap = (map: mapboxgl.Map, tileServer: string) => {
    map.addSource("public.points_odometer", {
      type: "vector",
      tiles: [tileServer],
    });
    map.addLayer({
      id: "public.points_odometer",
      type: "circle",
      source: "public.points_odometer",
      "source-layer": "default",
      paint: {
        "circle-radius": 3,
        "circle-color": "#f5efbc",
        "circle-stroke-color": "#f5efbc",
        "circle-stroke-width": 1,
        "circle-opacity": 0.5,
      },
    });

    map.addSource("public.trips_odometer", {
      type: "vector",
      tiles: [tileServer],
    });
    map.addLayer({
      id: "public.trips_odometer",
      type: "line",
      source: "public.trips_odometer",
      "source-layer": "default",
      layout: {
        "line-join": "round",
        "line-cap": "round",
      },
      paint: {
        "line-color": "#43d1d1",
        "line-width": 3,
      },
    });

    map.addSource("public.trips_speed", {
      type: "vector",
      tiles: [tileServer],
    });
    map.addLayer({
      id: "public.trips_speed",
      type: "line",
      source: "public.trips_speed",
      "source-layer": "default",
      layout: {
        "line-join": "round",
        "line-cap": "round",
      },
      paint: {
        "line-color": "#fc5e03",
        "line-width": 3,
      },
    });

    map.on("mousemove", "public.points_odometer", () => {
      map.getCanvas().style.cursor = "pointer";
    });

    // map.on("click", "public.trips_odometer", (event) => {
    //   setMapCenter({
    //     lat: event.lngLat.lat,
    //     lon: event.lngLat.lng,
    //     zoom: map.getZoom(),
    //   });
    //   // eslint-disable-next-line
    //   setSelectedDevice(event.features![0].properties!.devicekey as string);
    // });
  };

  return (
    <div className="App">
      <h1>DIMO User Trips 🚗</h1>
      <div>
        <div className="side-by-side">
          <button
          // onClick={() => {
          //   toggleLayer("public.trips_odometer");
          // }}
          >
            Sign In
          </button>
        </div>
        <br />
      </div>
      <div className="Container">
        <MapBox
          key={mapkey}
          onStyleLoad={(map) => {
            addHeatMap(map, tileserver);
            setMyMap(map);
          }}
          zoom={mapCenter.zoom ? [mapCenter.zoom] : [2.5]}
          center={
            mapCenter.lat
              ? [mapCenter.lon, mapCenter.lat]
              : [-50.200489, 37.38948]
          }
          // eslint-disable-next-line
          style={styleUrl}
          containerStyle={{
            height: "45vh",
            width: "100%",
          }}
        ></MapBox>
        <div className="Table">
          <div className="DropDown">
            <h2>Selected Device: {selectedDevice}</h2>
            <br />
            <label htmlFor="DeviceIDs"> Select a Device ID:</label>
            <select
              name="DeviceIDs"
              id="DeviceIDs"
              value={selectedDevice}
              onChange={(e) => {
                const data = JSON.parse(e.target.value) as string;
                // setMapCenter({ lat: data.lat, lon: data.lon, zoom: 10 });
                setSelectedDevice(data);
              }}
            >
              <option key="" value="">
                Select a device
              </option>
              {deviceIDs.map((device) => (
                <option key={device} value={JSON.stringify(device)}>
                  {device}
                </option>
              ))}
            </select>
            <div className="side-by-side">
              {/* <Table
                selectedDevice={selectedDevice}
                baseURL="http://localhost:8000/odometertrip/"
                title="Trips Calculated Using Odometer"
              />
              <div style={{ width: "10px" }}></div> */}
              <Table
                selectedDevice={selectedDevice}
                baseURL="http://localhost:8000"
                title="Trips Calculated Using Speed"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

interface TableProps {
  selectedDevice: string;
  baseURL: string;
  title: string;
}

const Table = ({ selectedDevice, baseURL, title }: TableProps) => {
  const [deviceInfo, setDeviceInfo] = useState<DeviceTrips[]>([]);
  useEffect(() => {
    if (selectedDevice !== "") {
      fetch(`${baseURL}/devices/${selectedDevice}/alltrips`, {
        method: "GET",
        headers: new Headers({
          Authorization:
            "Bearer " +
            "eyJhbGciOiJSUzI1NiIsImtpZCI6IjU1ZDhlMDU2MzBmMmI1OGFhMjQzNTVlOTM2OWE1OGQ3NzU3MjFmMzMifQ.eyJpc3MiOiJodHRwOi8vMTI3LjAuMC4xOjU1NTYvZGV4IiwicHJvdmlkZXJfaWQiOiJtb2NrIiwic3ViIjoiQ2cwd0xUTTROUzB5T0RBNE9TMHdFZ1J0YjJOciIsImF1ZCI6ImV4YW1wbGUtYXBwIiwiZXhwIjoxNjYyNzk4NTk1LCJpYXQiOjE2NjI3NTUzOTUsImF0X2hhc2giOiJ2OWhrVkwtLUR6cHRaZndJUXFuWkVRIiwiY19oYXNoIjoiR2VYNzN3Nk9PRC0tb1JVejcwcGlvUSIsImVtYWlsIjoia2lsZ29yZUBraWxnb3JlLnRyb3V0IiwiZW1haWxfdmVyaWZpZWQiOnRydWUsIm5hbWUiOiJLaWxnb3JlIFRyb3V0In0.syvb9LQdfYoTwRXmlfp8siokzu5pt6QTNgXWfxJV8AyvLNhp7flbMe0NXioPlOOtJv4m-rZdU6cx3wDqk0h3RGCJR-GHABJfjF7djiFNI14Hk_Ox2wceQxUYpiXkhGDLYANfBBy-atDTlwdHtar2Vz-zyBogvwcoL-tB9toXZCxbLulmzLfPDndTsHFOoOzPiLul1Xn_OGdX2rxASyXb99Ed0heGk3kJ1jicEqhDEA8cBGGvxq-7k04WlVLtXZTv-VhCbRud4cfnfef8-3eS8kkRYiyPnNLIVTq8-b0qeyRVXUkB_hdY_ewuN_ke9IyncL-nKWKS8DyO3Od74d8cVQ",
        }),
      })
        .then((r) => r.json())
        .then((data) => setDeviceInfo(data as []))
        .catch((e) => console.error(e));
    }
  }, [selectedDevice]);
  return (
    <div>
      {" "}
      <h3>{title}</h3>
      <br />
      <table>
        {selectedDevice !== "" &&
          deviceInfo.map((info, i) => {
            return (
              <tr key={i}>
                <th>{info.trip_id}</th>
                <th>{info.device_id}</th>
                <th>{info.trip_start}</th>
                <th>{info.trip_end}</th>
              </tr>
            );
          })}
      </table>
    </div>
  );
};

export default App;

// function MockSignIn() {
//   fetch(`http://127.0.0.1:5555/callback?code=iccs65ckpjmspfzeap3wlewbv&state=I+wish+to+wash+my+irish+wristwatch`, {
//     method: "GET"
//   })
//     .then((r) => {
//       return r.json();
//     })

//   url =
//     "http://127.0.0.1:5555/callback?code=iccs65ckpjmspfzeap3wlewbv&state=I+wish+to+wash+my+irish+wristwatch";
// }
