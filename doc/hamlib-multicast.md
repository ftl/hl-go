# Hamlib Multicast Feature — Developer Documentation

## 1. Overview

Hamlib's multicast feature allows the library (typically through `rigctld`) to **broadcast rig state information over the network using UDP multicast**. This enables multiple client applications to passively receive real-time radio status updates — frequency, mode, PTT state, split configuration, spectrum data, and more — without each application having to individually poll the rig through a serial port or TCP connection.

### Why Multicast?

Traditional Hamlib operation uses a **request/response model** over TCP (via `rigctld`). Each client must connect, send a command like `\get_freq`, and wait for a reply. This creates several problems:

- **Polling overhead**: Multiple applications (logger, digital modes software, spectrum display) all hammering the rig with get commands creates contention and latency.
- **Single-client bottleneck**: While `rigctld` does support multiple TCP clients, heavy polling from many clients can slow things down.
- **No push notifications**: The rig may change state (user turns the VFO knob) and clients won't know until they next poll.

Multicast solves this with a **publish/subscribe model**: Hamlib periodically broadcasts a JSON snapshot of the entire rig state to a multicast group. Any number of clients can listen passively. No polling needed.

### Feature Status

- **Planned milestone**: Hamlib 4.6 (originally discussed for 4.3)
- **Current state**: The multicast publisher (server-side broadcast) is implemented and functional. The multicast receiver (for sending commands back to the rig via multicast) exists in the codebase but is not fully implemented yet.
- **Primary use**: One-way broadcast of rig state from Hamlib → clients. Bidirectional control is a future goal.

---

## 2. Protocol Details

### Transport Layer

| Property              | Value                                      |
|-----------------------|--------------------------------------------|
| **Protocol**          | UDP (User Datagram Protocol)               |
| **Delivery method**   | IP Multicast                               |
| **Default address**   | `224.0.0.1` (link-local multicast)         |
| **Default port**      | `4532` (same as rigctld TCP default)       |
| **Data format**       | JSON (UTF-8 encoded text)                  |
| **Packet structure**  | One complete JSON document per UDP datagram |

Each UDP datagram contains a **single, self-contained JSON object** representing a full snapshot of the rig's current state. There is no fragmentation or multi-packet assembly — each packet is independently parseable.

### Multicast Address Space

The default multicast address `224.0.0.1` is in the **link-local** range (`224.0.0.0/24`), meaning packets are not forwarded by routers and stay on the local network segment. This is appropriate for a ham shack where all applications run on the same machine or LAN.

You can configure a different multicast group address (e.g., `239.x.x.x` for organization-local scope) via the `--multicast-addr` option.

---

## 3. JSON Packet Format

### Top-Level Structure

```json
{
  "app": "Hamlib",
  "version": "4.6~git 2023-10-20T16:52:59Z SHA=d87671",
  "seq": 183,
  "time": "2023-10-20T20:13:53.139869-0000",
  "crc": 0,
  "rig": { ... },
  "vfos": [ ... ],
  "spectra": [ ... ]
}
```

### Top-Level Fields

| Field       | Type    | Description |
|-------------|---------|-------------|
| `app`       | string  | Always `"Hamlib"` — identifies the source application. |
| `version`   | string  | Hamlib version string including git SHA and build date. |
| `seq`       | integer | Monotonically increasing 32-bit sequence number (1-up). Wraps from 2³²−1 back to 1. Use this to detect missed packets or reordering. |
| `time`      | string  | ISO 8601 timestamp of when the packet was generated. |
| `crc`       | integer | 32-bit CRC of the entire JSON record (with the CRC field set to `0` during computation). Currently `0` (not yet implemented in all builds). |

### The `rig` Object

```json
"rig": {
  "id": "FLRig:127.0.0.1:12345:30508",
  "status": "OK",
  "errorMsg": "",
  "name": "FLRig",
  "split": false,
  "splitVfo": "VFOA",
  "satMode": false,
  "modelist": "AM CW USB LSB FM CWR PKTLSB PKTUSB"
}
```

| Field       | Type    | Description |
|-------------|---------|-------------|
| `id`        | string  | Unique identifier for this rig instance. Format varies: may include model, endpoint, and process ID (e.g., `"IC-7300:com1:53535"` or `"FLRig:127.0.0.1:12345:30508"`). Allows distinguishing multiple rigs broadcasting on the same multicast group. |
| `status`    | string  | `"OK"`, `"Offline"`, or `"Error"`. |
| `errorMsg`  | string  | Empty when status is OK; contains error description otherwise. |
| `name`      | string  | Human-readable rig model name (e.g., `"IC-7300"`, `"FLRig"`, `"Dummy"`). |
| `split`     | boolean | Whether split operation is enabled. |
| `splitVfo`  | string  | Which VFO is used for the split TX frequency. |
| `satMode`   | boolean | Whether satellite mode is active. |
| `modelist`  | string  | Space-separated list of modes the rig supports. |

### The `vfos` Array

Each element describes one VFO:

```json
"vfos": [
  {
    "name": "VFOA",
    "freq": 14074270,
    "mode": "PKTUSB",
    "width": 3,
    "ptt": false,
    "rx": true,
    "tx": true
  },
  {
    "name": "VFOB",
    "freq": 14074500,
    "mode": "",
    "width": 0,
    "ptt": false,
    "rx": false,
    "tx": false
  }
]
```

| Field   | Type    | Description |
|---------|---------|-------------|
| `name`  | string  | VFO identifier: `"VFOA"`, `"VFOB"`, `"Main"`, `"Sub"`, etc. |
| `freq`  | integer | Frequency in Hz. |
| `mode`  | string  | Operating mode token (e.g., `"USB"`, `"LSB"`, `"CW"`, `"FM"`, `"PKTUSB"`). Empty string if unknown. |
| `width` | integer | Passband filter width in Hz. `0` if unknown. |
| `ptt`   | boolean | Whether PTT is active on this VFO. |
| `rx`    | boolean | Whether this VFO is currently receiving. |
| `tx`    | boolean | Whether this VFO is currently transmitting (or designated for TX). |

### The `spectra` Array (Optional)

Present when the rig supports spectrum/FFT data output (e.g., IC-7300):

```json
"spectra": [
  {
    "id": 0,
    "name": "Main",
    "type": "CENTER",
    "minLevel": 0,
    "maxLevel": 160,
    "minStrength": -80,
    "maxStrength": 0,
    "centerFreq": 3718000,
    "span": 50000,
    "lowFreq": 3693000,
    "highFreq": 3743000,
    "length": 475,
    "data": "121514000000000000000811..."
  }
]
```

| Field         | Type    | Description |
|---------------|---------|-------------|
| `id`          | integer | Numeric ID for this spectrum stream (matches rig caps). |
| `name`        | string  | Human name, typically corresponds to a VFO (`"Main"`, `"Sub"`). |
| `type`        | string  | `"FIXED"` or `"CENTER"` — whether the spectrum window is fixed or centered on the current frequency. |
| `minLevel` / `maxLevel` | integer | Raw FFT level range from the rig. |
| `minStrength` / `maxStrength` | integer | Signal strength range in dB. |
| `centerFreq`  | integer | Center frequency of the spectrum window in Hz. |
| `span`        | integer | Total frequency span in Hz. |
| `lowFreq` / `highFreq` | integer | Calculated low/high edges of the spectrum window in Hz. |
| `length`      | integer | Number of FFT data bytes. |
| `data`        | string  | Hex-encoded FFT data — each byte as 2 hex chars, so string length = 2 × `length`. |

### The `lastCommand` Object (Planned)

```json
"lastCommand": {
  "id": "MyApp 123",
  "command": "set_freq VFOA 14074000",
  "status": "OK"
}
```

This will echo the last command received by the rig, allowing clients to see each other's actions. The `id` is recommended to be an application name plus a sequence number.

---

## 4. Server Configuration (rigctld)

The multicast publisher is started automatically by Hamlib when a rig is opened. With `rigctld`, the relevant command-line options are:

### Key Options

| Option | Long Form | Description |
|--------|-----------|-------------|
| `-M`   | `--multicast-addr=ADDR` | Set the multicast group address. Default: `224.0.0.1`. Set to `0.0.0.0` to disable multicast. |
|        | `--multicast-port=PORT` | Set the UDP port for multicast. Default: `4532`. |

### Example: Basic Setup

```bash
# Start rigctld with an IC-7300, multicast enabled (defaults)
rigctld --model=3073 --rig-file=/dev/ttyUSB0 --serial-speed=115200
```

By default, Hamlib will start the multicast publisher automatically. The publisher will begin broadcasting rig state snapshots to `224.0.0.1:4532`.

### Example: Custom Multicast Address and Port

```bash
rigctld --port=20001 \
        --model=3073 \
        --serial-speed=115200 \
        --rig-file=/dev/ttyUSB0 \
        --set-conf=rts_state=OFF \
        --set-conf=dtr_state=OFF \
        --multicast-addr=224.0.0.1 \
        --multicast-port=20001 \
        --set-conf=async=1
```

### Example: Disabling Multicast

If the multicast address is set to `0.0.0.0`, the publisher will not start:

```bash
rigctld --model=3073 --rig-file=/dev/ttyUSB0 --multicast-addr=0.0.0.0
```

### Enabling Async + Spectrum Data

For rigs that support asynchronous data (e.g., Icom rigs), enable async mode and spectrum output:

```bash
rigctld --model=3073 --rig-file=/dev/ttyUSB0 --set-conf=async=1
```

Then, from a rigctl client or via TCP, enable spectrum output:

```
\set_func SPECTRUM 1
```

---

## 5. Receiving Multicast Data (Client Implementation)

### How UDP Multicast Reception Works

To receive multicast UDP packets, a client must:

1. **Create a UDP socket** (`SOCK_DGRAM`)
2. **Enable address reuse** (`SO_REUSEADDR` / `SO_REUSEPORT`) so multiple clients on the same machine can all receive
3. **Bind** the socket to `INADDR_ANY` on the multicast port (e.g., `4532`)
4. **Join the multicast group** using `IP_ADD_MEMBERSHIP` with the group address (e.g., `224.0.0.1`) and the local interface
5. **Read datagrams** in a loop — each one is a complete JSON packet

### Go Example Client
```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// --- JSON packet structures matching the Hamlib multicast schema ---

type HamlibPacket struct {
	App     string        `json:"app"`
	Version string        `json:"version"`
	Seq     uint32        `json:"seq"`
	Time    string        `json:"time"`
	CRC     uint32        `json:"crc"`
	Rig     RigInfo       `json:"rig"`
	VFOs    []VFO         `json:"vfos"`
	Spectra []SpectrumData `json:"spectra"`
}

type RigInfo struct {
	ID       interface{} `json:"id"`       // string or object depending on version
	Status   string      `json:"status"`
	ErrorMsg string      `json:"errorMsg"`
	Name     string      `json:"name"`
	Split    bool        `json:"split"`
	SplitVFO string      `json:"splitVfo"`
	SatMode  bool        `json:"satMode"`
	ModeList string      `json:"modelist"`
}

type VFO struct {
	Name  string  `json:"name"`
	Freq  uint64  `json:"freq"`
	Mode  string  `json:"mode"`
	Width int     `json:"width"`
	PTT   bool    `json:"ptt"`
	RX    bool    `json:"rx"`
	TX    bool    `json:"tx"`
}

type SpectrumData struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	MinLevel    int     `json:"minLevel"`
	MaxLevel    int     `json:"maxLevel"`
	MinStrength int     `json:"minStrength"`
	MaxStrength int     `json:"maxStrength"`
	CenterFreq  float64 `json:"centerFreq"`
	Span        int     `json:"span"`
	LowFreq     float64 `json:"lowFreq"`
	HighFreq    float64 `json:"highFreq"`
	Length      int     `json:"length"`
	Data        string  `json:"data"`
}

// --- Configuration ---

const (
	MulticastGroup = "224.0.0.1"
	MulticastPort  = 4532
	MaxPacketSize  = 65535
)

func main() {
	addr := fmt.Sprintf("%s:%d", MulticastGroup, MulticastPort)

	// Resolve the multicast group address
	groupAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		log.Fatalf("Failed to resolve multicast address %s: %v", addr, err)
	}

	// Join the multicast group on all interfaces.
	// ListenMulticastUDP handles socket creation, SO_REUSEADDR,
	// binding, and IP_ADD_MEMBERSHIP in one call.
	conn, err := net.ListenMulticastUDP("udp4", nil, groupAddr)
	if err != nil {
		log.Fatalf("Failed to join multicast group: %v", err)
	}
	defer conn.Close()

	// Set a generous read buffer so the OS doesn't drop packets
	// under burst conditions (e.g. fast spectrum data).
	if err := conn.SetReadBuffer(MaxPacketSize * 4); err != nil {
		log.Printf("Warning: could not set read buffer size: %v", err)
	}

	log.Printf("Listening for Hamlib multicast on %s ...", addr)

	// Graceful shutdown on Ctrl-C / SIGTERM
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nShutting down.")
		conn.Close()
		os.Exit(0)
	}()

	// --- Receive loop ---
	buf := make([]byte, MaxPacketSize)

	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		var pkt HamlibPacket
		if err := json.Unmarshal(buf[:n], &pkt); err != nil {
			log.Printf("JSON parse error (%d bytes from %s): %v", n, src, err)
			continue
		}

		printPacket(&pkt, src)
	}
}

func printPacket(pkt *HamlibPacket, src *net.UDPAddr) {
	fmt.Printf("\n--- Packet #%d from %s (src: %s) ---\n",
		pkt.Seq, pkt.Rig.Name, src.IP)
	fmt.Printf("  App: %s  Version: %s  Status: %s\n",
		pkt.App, pkt.Version, pkt.Rig.Status)

	if pkt.Rig.ErrorMsg != "" {
		fmt.Printf("  Error: %s\n", pkt.Rig.ErrorMsg)
	}

	for _, vfo := range pkt.VFOs {
		freqMHz := float64(vfo.Freq) / 1_000_000.0
		fmt.Printf("  %s: %.6f MHz  Mode=%-8s Width=%d Hz  RX=%t TX=%t PTT=%t\n",
			vfo.Name, freqMHz, vfo.Mode, vfo.Width, vfo.RX, vfo.TX, vfo.PTT)
	}

	fmt.Printf("  Split=%t  SplitVFO=%s  SatMode=%t\n",
		pkt.Rig.Split, pkt.Rig.SplitVFO, pkt.Rig.SatMode)

	for _, sp := range pkt.Spectra {
		if sp.Length > 0 {
			fmt.Printf("  Spectrum[%d] %q: center=%.0f Hz  span=%d Hz  points=%d\n",
				sp.ID, sp.Name, sp.CenterFreq, sp.Span, sp.Length)
		}
	}
}
```

A few things worth noting about the Go implementation:

Go's `net.ListenMulticastUDP` is quite convenient — it handles socket creation, `SO_REUSEADDR`, binding, and `IP_ADD_MEMBERSHIP` all in a single call, which is noticeably cleaner than the C or Python versions where you wire up each step manually.

The `nil` second argument (interface) means "join on all interfaces," equivalent to `INADDR_ANY`. If you need to pin it to a specific NIC (common on multi-homed machines), pass a `*net.Interface` from `net.InterfaceByName("eth0")` instead.

The `RigInfo.ID` field is typed as `interface{}` because older Hamlib versions send it as a plain string (e.g., `"FLRig:127.0.0.1:12345:30508"`) while the spec also shows a structured object form with `model`, `endpoint`, and `process` fields. Using `interface{}` lets `json.Unmarshal` handle either shape without failing.

To run it:

```bash
go run hamlib_multicast.go
```

You should see rig state packets printing as soon as `rigctld` is running with multicast enabled on the same network.

---

## 6. Multiple Rigs on the Same Multicast Group

The protocol supports multiple rigs broadcasting on the **same multicast address and port**. Each packet contains a `rig.id` field that uniquely identifies the rig instance. Clients should filter packets by this ID to track a specific rig.

The `rig.id` format encodes enough information to distinguish rigs: the model name, connection endpoint, and process ID. For example:

- `"FLRig:127.0.0.1:12345:30508"` — an FLRig connection with TCP endpoint and PID
- `"IC-7300:com1:53535"` — a direct serial connection with a PID

---

## 7. Design Principles for Client Parsers

The README.multicast file emphasizes several important guidelines:

1. **Allow for unknown fields**: New key-value pairs may be added to the JSON in future versions. Parsers should silently ignore fields they don't recognize rather than failing.

2. **Use the sequence number**: The `seq` field lets you detect missed packets and measure update rates. Since this is UDP, packets can be lost, duplicated, or arrive out of order (though on a local LAN, loss is rare).

3. **CRC validation** (future): When `crc` is nonzero, clients should validate the checksum by replacing the CRC value with `0` in the JSON string, computing the CRC32, and comparing. Currently CRC is typically `0` (unimplemented).

4. **Don't assume VFO count**: Some rigs have one VFO, others have two or more. The `vfos` array length will vary.

5. **Spectrum data is optional**: The `spectra` array may be empty or contain entries with `length: 0` if the rig doesn't support spectrum output or it hasn't been enabled.

---

## 8. Architecture Diagram

```
┌─────────────┐     Serial/USB/TCP      ┌──────────────────┐
│  Ham Radio   │◄───────────────────────►│     rigctld       │
│  (IC-7300,   │                         │                  │
│   FT-991A,   │                         │  ┌────────────┐  │
│   etc.)      │                         │  │  Multicast  │  │
└─────────────┘                         │  │  Publisher  │  │
                                         │  └──────┬─────┘  │
                                         │         │        │
                                         │  TCP    │ UDP    │
                                         │  :4532  │ Mcast  │
                                         └────┬────┼────────┘
                                              │    │
                          ┌───────────────────┤    │
                          │                   │    │
                  ┌───────┴───────┐    ┌──────┴────▼──────┐
                  │  TCP Client   │    │  Multicast Client │
                  │  (rigctl,     │    │  (any language,   │
                  │   WSJT-X,     │    │   passive listen) │
                  │   logging)    │    │                   │
                  │               │    │  ┌─────────────┐  │
                  │  Request/     │    │  │ UDP Socket   │  │
                  │  Response     │    │  │ Join Group   │  │
                  │  model        │    │  │ 224.0.0.1    │  │
                  └───────────────┘    │  │ Port 4532    │  │
                                       │  └─────────────┘  │
                                       └──────────────────┘
```

**TCP clients** (existing model): send commands, receive responses. One-at-a-time request/response.

**Multicast clients** (new model): passively receive continuous rig state broadcasts. Zero load on the rig. Any number of simultaneous listeners.

---

## 9. Comparison: TCP vs. Multicast

| Aspect              | TCP (rigctld)                    | UDP Multicast                        |
|---------------------|----------------------------------|--------------------------------------|
| **Direction**       | Bidirectional (request/response) | Primarily broadcast (server → clients) |
| **Connection**      | Persistent TCP connection        | Connectionless (join group)          |
| **Client overhead** | Must poll for updates            | Passive — data pushed automatically  |
| **Multiple clients**| Each needs its own connection    | Unlimited listeners on the group     |
| **Reliability**     | TCP guarantees delivery          | UDP — packets may be lost            |
| **Latency**         | Depends on poll interval         | Near real-time (broadcast interval)  |
| **Rig control**     | Full command set                 | Read-only (control is future work)   |
| **Protocol**        | Text commands (rigctl protocol)  | JSON over UDP                        |

---

## 10. Troubleshooting

**No packets received?**

- Verify `rigctld` is running with a non-zero multicast address (not `0.0.0.0`).
- Check debug output: run rigctld with `-vvvvv` and look for log lines containing `network_multicast_publisher_start`. If you see `"not starting multicast publisher"`, the address is set to `0.0.0.0`.
- Ensure your firewall allows UDP traffic on the multicast port.
- On multi-homed machines, you may need to specify the correct network interface when joining the multicast group (replace `INADDR_ANY` with the specific interface address).
- On systems using IGMPv3 (Source Specific Multicast), you may need to explicitly specify the source IP or downgrade to IGMPv2.

**Garbled or partial data?**

- Each UDP datagram should contain one complete JSON object. If you're seeing partial JSON, your receive buffer may be too small — use at least 65535 bytes.
- Check the `seq` field for gaps indicating missed packets.

**Multiple rigs conflicting?**

- Filter incoming packets by `rig.id` to isolate specific rigs.
- Alternatively, run rigs on different multicast ports using `--multicast-port`.

---

## 11. References

- **README.multicast**: [github.com/Hamlib/Hamlib/blob/master/README.multicast](https://github.com/Hamlib/Hamlib/blob/master/README.multicast) — the canonical protocol specification
- **Issue #695**: [github.com/Hamlib/Hamlib/issues/695](https://github.com/Hamlib/Hamlib/issues/695) — the feature request and discussion
- **Hamlib NEWS file**: [github.com/Hamlib/Hamlib/blob/master/NEWS](https://github.com/Hamlib/Hamlib/blob/master/NEWS) — release notes documenting multicast progress
- **Hamlib Wiki**: [github.com/Hamlib/Hamlib/wiki](https://github.com/Hamlib/Hamlib/wiki) — general project documentation
- **Source code**: `src/network.c` contains `network_multicast_publisher_start`, `network_multicast_publisher_stop`, `network_multicast_receiver_start`, and `network_multicast_receiver_stop`
