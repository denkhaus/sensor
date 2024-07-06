# sensor

This code reads data from a `CWT-Soil-THC-S` soil sensor and sends the readings by mqtt. The following parameters will be processed:

```yaml {"id":"01J24JD3PVE2Z7A66EXCR04N18"}
- Humidity:     01 03 00 00 00 01 84 0a
- Temperatur:   01 03 00 01 00 01 d5 ca
- Conductivity: 01 03 00 02 00 01 25 ca
- Salinity:     01 03 00 03 00 01 74 0a
- TDS:          01 03 00 04 00 01 c5 cb
```

### scripting

Additionally you can react on sensor data by scripting. Please see [sensor_script.go](https://github.com/denkhaus/sensor/blob/master/sensor_script.go) for more information. The script is periodically executed by yaegi interpreter. So your'e able to switch pumps, valves etc. based on sensor data.

This is work in progress. The code will change without further notice.
