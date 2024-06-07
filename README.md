# sensor

This code reads data from a `CWT-Soil-THC-S` soil sensor and sends it by mqtt. The following parameters will be processed:

```
- Humidity:     01 03 00 00 00 01 84 0a
- Temperatur:   01 03 00 01 00 01 d5 ca
- Conductivity: 01 03 00 02 00 01 25 ca
- Salinity:     01 03 00 03 00 01 74 0a
- TDS:          01 03 00 04 00 01 c5 cb
```
