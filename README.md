# sonoff-lan
Sonoff Lan Mode Tools

A few hacks to try and figure out Sonof "LAN Mode" devices. 

- Forc devices into LAN mode
- How to discover them
- How to then manage them

## Forcing LAN mode
Before forcing a Sonof device into LAN mode, it needs to be provisioned/paired as per normal using the eWeLink APP. This is important. If not done, you won't be able to see the device later in the eWelink APP, in LAN mode. 

If the ability to manage the device in LAN mode via the eWeLink App is not important to you,then carry on without the eWeLink app, to next section: "Push the device button"...

After provisioning it normally using the eWelink App, a Sonof device can be perpetually forced into lan mode,by resetting the device to defaults and provisioning it against a non-existent management server.

#### Push the device button
Push the button for 7 seconds and if required, release and push for another 7 seconds until the LED starts blinking rapidly. Some devices don't go into provisioning mode immediately. The LED needs to flash rapidly, with no pattern. Push the button for 7 seconds, release, and try again until you see the LED blinking rapidly (with no pattern)

Once the LED is blinking rapidly, the device will turn into a WiFi Access Point, and you should see a new SSID.

#### Connect to the device via WiFi
If you are successful in pushing the button, and getting the device into provisioning mode, you will see a WiFi SSID "IDEAT-XXXX". 
If you do NOT, then try again.

Connect to "DEAT-XXXXX" WiFi network  with your computer, the password is "12345678", and you will get assigned a DHCP address from the Sonof device.

Check if you can ping 10.10.7.1
```
ping 10.10.7.1
```

If you can ping it, the device is now ready to be forced to "LAN Mode"

#### Run the "FORCE LAN MODE" provisioning command

Run the following curl command:

```
curl -v -v -d "{ \"ssid\": \"YOURWIFI_SSID\", \"version\": 4, \"password\": \"YOURPASSWORD\", \"serverName\": \"10.99.99.99\", \"port\": 8443 }" -X POST http://10.10.7.1/ap 
```

The device is now provisioned to connect to an IOT server 10.99.99.99:8443, which will hopefully never be reachable on your LAN network. If it is, just change it to something random.

The device should register on your WiFi after a few moments, and you should be able to manage it with the eWelink Android/IOS application by enabling "LAN mode" in the App. If it doesn't appear, then you have a problem...

## Discovery

The 
