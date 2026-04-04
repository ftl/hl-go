# Hamlib Extended Response Protocol — Developer Reference

_A comprehensive reference for implementing a Go client library for rigctld, rotctld, and ampctld._

---

## 1. Architecture Overview

Hamlib provides three TCP daemons that expose hardware control over the network:

| Daemon    | Controls   | Default Port | NET Model ID |
| --------- | ---------- | ------------ | ------------ |
| `rigctld` | Radios     | 4532         | 2            |
| `rotctld` | Rotators   | 4533         | 2            |
| `ampctld` | Amplifiers | 4531         | 2            |

All three share the same wire protocol design: a line-oriented text protocol over a persistent TCP connection. The client sends one command per line, terminated by `\n`. The server responds with one or more lines, also `\n`-terminated.

There are two protocol variants: the **Default Protocol** (simple, used by the NET backend models for library-to-daemon communication) and the **Extended Response Protocol** (structured, recommended for third-party clients). This document focuses on the Extended Response Protocol since it provides deterministic parsing.

---

## 2. Transport Layer

### Connection

Plain TCP. No TLS. No authentication (though a password mechanism via `-A` is planned for rigctld). Clients open a standard TCP socket to `host:port`.

```go
conn, err := net.Dial("tcp", "localhost:4532")
```

### Framing

The protocol is **line-oriented**. Each line is terminated by `\n` (0x0A). There is no length prefix, no binary framing. A single command/response exchange may span multiple lines — the `RPRT` line signals the end of a response block.

### Encoding

All data is **ASCII text**. Numeric values are decimal integers or floating-point strings. There are no binary payloads in the TCP protocol.

### Connection Lifecycle

The TCP connection is persistent. A client connects once and can send an arbitrary number of commands sequentially. There is no explicit session handshake or login. The connection remains open until the client disconnects or `rigctld` shuts down. Multiple clients can connect simultaneously but may experience contention.

---

## 3. The Two Protocols

### 3.1 Default Protocol

Used internally by Hamlib's NET backend (model 2). Simple but harder to parse programmatically.

**Sending a command:**

```
F 14250000\n
```

**Response to a set command (success):**

```
RPRT 0\n
```

**Response to a set command (error):**

```
RPRT -1\n
```

**Response to a get command:**

```
14250000\n
```

The default protocol returns raw values with no labels. Error responses use the `RPRT x` format where `x` is the negated Hamlib error code. Get commands return only values (no keys), and the number of returned lines varies per command.

### 3.2 Extended Response Protocol (Recommended)

The Extended Response Protocol (ERP) adds structure that makes parsing deterministic. **This is the protocol you should implement for a client library.**

ERP is activated by **prepending a punctuation character** to each command. This character determines the record separator used in the response.

---

## 4. Extended Response Protocol — Detailed Specification

### 4.1 Activating ERP

Prepend one of the following characters to the command string:

| Prefix | Separator Behavior                           | Recommended For                            |
| ------ | -------------------------------------------- | ------------------------------------------ |
| `+`    | Each record on its own line (`\n`-separated) | General use, easiest to parse line-by-line |
| `;`    | All records on one line, separated by `;`    | Single-event reads                         |
| `\|`   | All records on one line, separated by `\|`   | Pipe-delimited parsing                     |
| `,`    | All records on one line, separated by `,`    | CSV-style parsing                          |

**Reserved characters that must NOT be used as prefixes:**

| Char | Reason                                        |
| ---- | --------------------------------------------- |
| `\`  | Used for long command name prefix             |
| `?`  | Reserved for help/query in interactive mode   |
| `_`  | Reserved for `get_info` short command         |
| `#`  | Reserved for comments in command file scripts |

Any other C `ispunct()` character may work as a separator but is untested.

**Recommendation for Go clients:** Use `+` exclusively. It produces multi-line responses that are trivial to parse with `bufio.Scanner`.

### 4.2 Response Structure

Every ERP response follows exactly this structure:

```
<command_echo><sep><key>: <value><sep>...<key>: <value><sep>RPRT <code><sep>
```

Formally, there are four rules:

**Rule 1 — Command Echo:** The first record is always the long command name, followed by any arguments that were sent. This echoes what the server received.

**Rule 2 — End Marker:** The last record is always `RPRT x` where `x` is the numeric return code (`0` = success, negative = error).

**Rule 3 — Data Records:** Any data returned by the backend appears as `Key: Value` pairs between the command echo and the RPRT line. The key name is followed by a colon, a space, and then the value.

**Rule 4 — Acknowledgement Guarantee:** Every command, whether set or get, always produces at least the command echo (Rule 1) and the RPRT line (Rule 2). Data records (Rule 3) only appear when the command returns values.

### 4.3 Examples with `+` Separator

**Set command (no return values):**

```
Client sends:  +F 14250000\n
Server returns:
  set_freq: 14250000\n
  RPRT 0\n
```

**Get command (returns values):**

```
Client sends:  +f\n
Server returns:
  get_freq:\n
  Frequency: 14250000\n
  RPRT 0\n
```

**Get command returning multiple values:**

```
Client sends:  +m\n
Server returns:
  get_mode:\n
  Mode: USB\n
  Passband: 2400\n
  RPRT 0\n
```

**Error response:**

```
Client sends:  +F invalid\n
Server returns:
  set_freq: invalid\n
  RPRT -1\n
```

### 4.4 Examples with `;` Separator (Single-Line)

```
Client sends:  ;f\n
Server returns:
  get_freq:;Frequency: 14250000;RPRT 0\n
```

```
Client sends:  ;m\n
Server returns:
  get_mode:;Mode: USB;Passband: 2400;RPRT 0\n
```

### 4.5 Parsing Algorithm (for `+` mode)

```
1. Read lines until you find one starting with "RPRT "
2. The first line is the command echo — discard or log it
3. The last line is "RPRT x" — parse x as the return code
4. All lines between are "Key: Value" data records
5. For each data line, split on the FIRST ": " to get key and value
```

**Important edge cases:**

- A command echo for a set command includes the arguments: `set_freq: 14250000`
- A command echo for a get command has no arguments: `get_freq:`
- The RPRT value is the **negated** Hamlib error code (0 = success, -1 = EINVAL, -4 = ENIMPL, etc.)

---

## 5. Command Reference

### 5.1 Command Syntax

Commands can be sent in two forms:

| Form                           | Example        | Description                                                                              |
| ------------------------------ | -------------- | ---------------------------------------------------------------------------------------- |
| Short (single char)            | `+f\n`         | Compact, case-sensitive. Uppercase = set, lowercase = get.                               |
| Long (backslash-prefixed name) | `+\get_freq\n` | Verbose, always lowercase with underscores. The `\` prefix is required even in ERP mode. |

Arguments are space-separated on the same line after the command character or name:

```
+F 14250000\n              — short form set
+\set_freq 14250000\n      — long form set (equivalent)
+M USB 2400\n              — short form with two arguments
```

### 5.2 VFO Mode

If `rigctld` was started with `--vfo` (or the client sends `\set_vfo_opt 1`), an extra VFO argument is required before each command's normal arguments:

```
+F VFOA 14250000\n         — with VFO mode enabled
```

You can detect VFO mode at runtime:

```
Client sends:  +\chk_vfo\n
Server returns:
  chk_vfo:\n
  VFO: 1\n                  — 1 means VFO mode is active
  RPRT 0\n
```

VFO tokens: `VFOA`, `VFOB`, `VFOC`, `currVFO`, `VFO`, `MEM`, `Main`, `Sub`, `TX`, `RX`, `MainA`, `MainB`, `SubA`, `SubB`.

### 5.3 rigctld Commands (Radio)

#### Frequency

| Command          | Short | Args                           | Response Keys  | Description                     |
| ---------------- | ----- | ------------------------------ | -------------- | ------------------------------- |
| `set_freq`       | `F`   | `Frequency` (Hz, int or float) | —              | Set VFO frequency               |
| `get_freq`       | `f`   | —                              | `Frequency`    | Get VFO frequency (Hz, integer) |
| `set_split_freq` | `I`   | `Tx Frequency`                 | —              | Set TX frequency for split      |
| `get_split_freq` | `i`   | —                              | `TX Frequency` | Get TX frequency                |

#### Mode

| Command          | Short | Args                    | Response Keys            | Description                         |
| ---------------- | ----- | ----------------------- | ------------------------ | ----------------------------------- |
| `set_mode`       | `M`   | `Mode` `Passband`       | —                        | Set operating mode and filter width |
| `get_mode`       | `m`   | —                       | `Mode`, `Passband`       | Get current mode and filter width   |
| `set_split_mode` | `X`   | `TX Mode` `TX Passband` | —                        | Set TX mode for split               |
| `get_split_mode` | `x`   | —                       | `TX Mode`, `TX Passband` | Get TX mode                         |

**Mode tokens:** `USB`, `LSB`, `CW`, `CWR`, `RTTY`, `RTTYR`, `AM`, `FM`, `WFM`, `AMS`, `PKTLSB`, `PKTUSB`, `PKTFM`, `ECSSUSB`, `ECSSLSB`, `FA`, `SAM`, `SAL`, `SAH`, `DSB`.

**Passband:** integer Hz, or `0` for backend default, or `-1` for no change.

**Query trick:** Sending `?` as the first argument returns a space-separated list of supported modes: `+M ?\n`

#### VFO

| Command   | Short | Args  | Response Keys |
| --------- | ----- | ----- | ------------- |
| `set_vfo` | `V`   | `VFO` | —             |
| `get_vfo` | `v`   | —     | `VFO`         |

#### Split

| Command         | Short | Args                   | Response Keys     |
| --------------- | ----- | ---------------------- | ----------------- |
| `set_split_vfo` | `S`   | `Split` (0/1) `TX VFO` | —                 |
| `get_split_vfo` | `s`   | —                      | `Split`, `TX VFO` |

#### PTT

| Command   | Short | Args  | Response Keys | Description                     |
| --------- | ----- | ----- | ------------- | ------------------------------- |
| `set_ptt` | `T`   | `PTT` | —             | 0=RX, 1=TX, 2=TX mic, 3=TX data |
| `get_ptt` | `t`   | —     | `PTT`         | Returns 0/1/2/3                 |

#### Tuning Offsets

| Command   | Short | Args                 | Response Keys |
| --------- | ----- | -------------------- | ------------- |
| `set_rit` | `J`   | `RIT` (Hz, ±integer) | —             |
| `get_rit` | `j`   | —                    | `RIT`         |
| `set_xit` | `Z`   | `XIT` (Hz, ±integer) | —             |
| `get_xit` | `z`   | —                    | `XIT`         |

#### Antenna

| Command   | Short | Args                           | Response Keys                         |
| --------- | ----- | ------------------------------ | ------------------------------------- |
| `set_ant` | `Y`   | `Antenna` (int) `Option` (int) | —                                     |
| `get_ant` | `y`   | `Antenna`                      | `AntCurr`, `Option`, `AntTx`, `AntRx` |

#### Functions (Boolean Rig Features)

| Command    | Short | Args                  | Response Keys |
| ---------- | ----- | --------------------- | ------------- |
| `set_func` | `U`   | `Func` `Status` (0/1) | —             |
| `get_func` | `u`   | `Func`                | `Func Status` |

**Function tokens:** `FAGC`, `NB`, `COMP`, `VOX`, `TONE`, `TSQL`, `SBKIN`, `FBKIN`, `ANF`, `NR`, `AIP`, `APF`, `MON`, `MN`, `RF`, `ARO`, `LOCK`, `MUTE`, `VSC`, `REV`, `SQL`, `ABM`, `BC`, `MBC`, `RIT`, `AFC`, `SATMODE`, `SCOPE`, `RESUME`, `TBURST`, `TUNER`, `XIT`, `SPECTRUM`.

#### Levels (Analog/Stepped Rig Parameters)

| Command     | Short | Args            | Response Keys |
| ----------- | ----- | --------------- | ------------- |
| `set_level` | `L`   | `Level` `Value` | —             |
| `get_level` | `l`   | `Level`         | `Level Value` |

**Level tokens:** `PREAMP`, `ATT`, `VOX`, `AF`, `RF`, `SQL`, `IF`, `APF`, `NR`, `PBT_IN`, `PBT_OUT`, `CWPITCH`, `RFPOWER`, `RFPOWER_METER`, `RFPOWER_METER_WATTS`, `MICGAIN`, `KEYSPD`, `NOTCHF`, `COMP`, `AGC`, `BKINDL`, `BAL`, `METER`, `VOXGAIN`, `ANTIVOX`, `SLOPE_LOW`, `SLOPE_HIGH`, `RAWSTR`, `SWR`, `ALC`, `STRENGTH`.

**AGC values:** 0=OFF, 1=SUPERFAST, 2=FAST, 3=SLOW, 4=USER, 5=MEDIUM, 6=AUTO.

#### Parameters (Rig Configuration)

| Command    | Short | Args           | Response Keys |
| ---------- | ----- | -------------- | ------------- |
| `set_parm` | `P`   | `Parm` `Value` | —             |
| `get_parm` | `p`   | `Parm`         | `Parm Value`  |

**Parm tokens:** `ANN`, `APO`, `BACKLIGHT`, `BEEP`, `TIME`, `BAT`, `KEYLIGHT`, `SCREENSAVER`, `BANDSELECT`.

#### Repeater

| Command          | Short | Args                    | Response Keys |
| ---------------- | ----- | ----------------------- | ------------- |
| `set_rptr_shift` | `R`   | `Rptr Shift` (+/-/None) | —             |
| `get_rptr_shift` | `r`   | —                       | `Rptr Shift`  |
| `set_rptr_offs`  | `O`   | `Rptr Offset` (Hz)      | —             |
| `get_rptr_offs`  | `o`   | —                       | `Rptr Offset` |

#### Tone/CTCSS/DCS

| Command          | Short  | Args                     | Response Keys |
| ---------------- | ------ | ------------------------ | ------------- |
| `set_ctcss_tone` | `C`    | `CTCSS Tone` (tenths Hz) | —             |
| `get_ctcss_tone` | `c`    | —                        | `CTCSS Tone`  |
| `set_dcs_code`   | `D`    | `DCS Code`               | —             |
| `get_dcs_code`   | `d`    | —                        | `DCS Code`    |
| `set_ctcss_sql`  | `0x90` | `CTCSS Sql` (tenths Hz)  | —             |
| `get_ctcss_sql`  | `0x91` | —                        | `CTCSS Sql`   |
| `set_dcs_sql`    | `0x92` | `DCS Sql`                | —             |
| `get_dcs_sql`    | `0x93` | —                        | `DCS Sql`     |

#### Memory

| Command       | Short | Args                            | Response Keys             |
| ------------- | ----- | ------------------------------- | ------------------------- |
| `set_mem`     | `E`   | `Memory#`                       | —                         |
| `get_mem`     | `e`   | —                               | `Memory#`                 |
| `set_bank`    | `B`   | `Bank`                          | —                         |
| `set_channel` | `H`   | `Channel` (not yet implemented) | —                         |
| `get_channel` | `h`   | `readonly` (0/1)                | (multi-line channel data) |

#### VFO Operations

| Command  | Short | Args                      |
| -------- | ----- | ------------------------- |
| `vfo_op` | `G`   | `Mem/VFO Op`              |
| `scan`   | `g`   | `Scan Fct` `Scan Channel` |

**VFO Op tokens:** `CPY`, `XCHG`, `FROM_VFO`, `TO_VFO`, `MCL`, `UP`, `DOWN`, `BAND_UP`, `BAND_DOWN`, `LEFT`, `RIGHT`, `TUNE`, `TOGGLE`.

#### Power & Status

| Command         | Short  | Args / Response Keys                              |
| --------------- | ------ | ------------------------------------------------- |
| `set_powerstat` | `0x87` | `Power Status` (0=Off, 1=On, 2=Standby)           |
| `get_powerstat` | `0x88` | Returns `Power Status`                            |
| `reset`         | `*`    | `Reset` (1=Software, 2=VFO, 4=MemClear, 8=Master) |

#### DTMF & Morse

| Command      | Short  | Args                             |
| ------------ | ------ | -------------------------------- |
| `send_morse` | `b`    | `Morse` (text string)            |
| `send_dtmf`  | `0x89` | `Digits`                         |
| `recv_dtmf`  | `0x8a` | Returns `Digits`                 |
| `get_dcd`    | `0x8b` | Returns `DCD` (0=Closed, 1=Open) |

#### Information & Introspection

| Command        | Short  | Description                                  |
| -------------- | ------ | -------------------------------------------- |
| `get_info`     | `_`    | Misc rig info string                         |
| `get_rig_info` | `0xf5` | Full rig state (VFOs, modes, etc.)           |
| `get_vfo_info` | `0xf3` | Info about a specific VFO (takes VFO arg)    |
| `dump_state`   | (none) | Backend state dump                           |
| `dump_caps`    | `1`    | Backend capabilities (many lines)            |
| `chk_vfo`      | (none) | Returns `CHKVFO 0` or `CHKVFO 1`             |
| `set_vfo_opt`  | (none) | `Status` (0/1) — dynamically toggle VFO mode |

#### Power Conversion

| Command    | Short | Args                            | Returns           |
| ---------- | ----- | ------------------------------- | ----------------- |
| `power2mW` | `2`   | `Power` (0.0–1.0) `Freq` `Mode` | `Power mW`        |
| `mW2power` | `4`   | `Power mW` `Freq` `Mode`        | `Power` (0.0–1.0) |

#### Lock Mode

| Command         | Args       | Description                                             |
| --------------- | ---------- | ------------------------------------------------------- |
| `set_lock_mode` | `0` or `1` | Prevents all clients from changing rig mode when locked |
| `get_lock_mode` | —          | Returns current lock state                              |

#### Separator & Password

| Command         | Args            | Description                                                   |
| --------------- | --------------- | ------------------------------------------------------------- |
| `set_separator` | `char`          | Change response separator dynamically (recommend `#` or `@`)  |
| `password`      | `shared_secret` | Authenticate with rigctld when `-A` is used (NOT IMPLEMENTED) |

### 5.4 rotctld Commands (Rotator)

| Command      | Short | Args                          | Response Keys          | Description                                                    |
| ------------ | ----- | ----------------------------- | ---------------------- | -------------------------------------------------------------- |
| `set_pos`    | `P`   | `Azimuth` `Elevation` (float) | —                      | Set rotator position                                           |
| `get_pos`    | `p`   | —                             | `Azimuth`, `Elevation` | Get position (float)                                           |
| `move`       | `M`   | `Direction` `Speed`           | —                      | Direction: 2=Up, 4=Down, 8=Left, 16=Right. Speed: 1–100 or -1. |
| `stop`       | `S`   | —                             | —                      | Stop rotation                                                  |
| `park`       | `K`   | —                             | —                      | Park rotator                                                   |
| `reset`      | `R`   | `Reset` (1=Reset All)         | —                      | Reset rotator                                                  |
| `get_info`   | `_`   | —                             | `Info`                 | Model name string                                              |
| `set_conf`   | `C`   | `Token` `Value`               | —                      | Set config parameter                                           |
| `dump_state` | —     | —                             | (multi-line)           | Backend state                                                  |
| `dump_caps`  | `1`   | —                             | (multi-line)           | Capabilities                                                   |

**Locator utility commands** (rotctld only):

| Command      | Short | Args                                | Returns                             |
| ------------ | ----- | ----------------------------------- | ----------------------------------- |
| `lonlat2loc` | `L`   | `Longitude` `Latitude` `Loc Len`    | `Locator` (Maidenhead)              |
| `loc2lonlat` | `l`   | `Locator`                           | `Longitude`, `Latitude`             |
| `dms2dec`    | `D`   | `Degrees` `Minutes` `Seconds` `S/W` | `Dec Degrees`                       |
| `dec2dms`    | `d`   | `Dec Degrees`                       | `Degrees` `Minutes` `Seconds` `S/W` |
| `dmmm2dec`   | `E`   | `Degrees` `Dec Minutes` `S/W`       | `Dec Degrees`                       |
| `dec2dmmm`   | `e`   | `Dec Deg`                           | `Degrees` `Minutes` `S/W`           |
| `qrb`        | `B`   | `Lon1` `Lat1` `Lon2` `Lat2`         | `Distance` (km), `Azimuth` (deg)    |
| `a_sp2a_lp`  | `A`   | `Short Path Deg`                    | `Long Path Deg`                     |
| `d_sp2d_lp`  | `a`   | `Short Path km`                     | `Long Path km`                      |

### 5.5 ampctld Commands (Amplifier)

| Command         | Short | Args             | Response Keys   | Description                            |
| --------------- | ----- | ---------------- | --------------- | -------------------------------------- |
| `set_freq`      | `F`   | `Frequency` (Hz) | —               | Set amp frequency                      |
| `get_freq`      | `f`   | —                | `Frequency(Hz)` | Get amp frequency                      |
| `get_level`     | `l`   | `Level`          | `Level Value`   | Get level (use `?` for supported list) |
| `set_powerstat` | —     | `Power Status`   | —               | 0=Off, 1=On, 2=Standby, 4=Operate      |
| `get_powerstat` | —     | —                | `Power Status`  | Get power state                        |
| `reset`         | `R`   | `Reset`          | —               | 0=None, 1=Memory, 2=Fault, 3=Amplifier |
| `get_info`      | `_`   | —                | `Info`          | Amp info                               |
| `dump_state`    | —     | —                | (multi-line)    | Backend state                          |
| `dump_caps`     | `1`   | —                | (multi-line)    | Capabilities                           |

---

## 6. Error Codes

The `RPRT` line returns `0` on success, or the **negated** value of a Hamlib error code on failure. The error codes are defined in `rig.h`:

| RPRT Value | Constant        | Meaning                                             |
| ---------- | --------------- | --------------------------------------------------- |
| `0`        | `RIG_OK`        | Success                                             |
| `-1`       | `RIG_EINVAL`    | Invalid parameter                                   |
| `-2`       | `RIG_ECONF`     | Invalid configuration (serial, etc.)                |
| `-3`       | `RIG_ENOMEM`    | Memory shortage                                     |
| `-4`       | `RIG_ENIMPL`    | Function not implemented (but may be in the future) |
| `-5`       | `RIG_ETIMEOUT`  | Communication timed out                             |
| `-6`       | `RIG_EIO`       | I/O error (including port open failure)             |
| `-7`       | `RIG_EINTERNAL` | Internal Hamlib error                               |
| `-8`       | `RIG_EPROTO`    | Protocol error                                      |
| `-9`       | `RIG_ERJCTED`   | Command rejected by the rig                         |
| `-10`      | `RIG_ETRUNC`    | Command performed but argument was truncated        |
| `-11`      | `RIG_ENAVAIL`   | Function not available                              |
| `-12`      | `RIG_ENTARGET`  | VFO not targetable                                  |
| `-13`      | `RIG_BUSERROR`  | Error talking on the bus                            |
| `-14`      | `RIG_BUSBUSY`   | Collision on the bus                                |
| `-15`      | `RIG_EARG`      | NULL handle or invalid pointer parameter            |
| `-16`      | `RIG_EVFO`      | Invalid VFO                                         |
| `-17`      | `RIG_EDOM`      | Argument out of domain of function                  |

"Soft" errors (non-fatal, informational) include: `-1`, `-4`, `-9`, `-10`, `-11`, `-12`, `-16`, `-17`.

---

## 7. Go Client Implementation Guide

### 7.1 Connection Management

```go
type Client struct {
    conn    net.Conn
    reader  *bufio.Reader
    mu      sync.Mutex   // serialize command/response pairs
    prefix  byte         // ERP prefix, typically '+'
}

func Dial(addr string) (*Client, error) {
    conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
    if err != nil {
        return nil, err
    }
    return &Client{
        conn:   conn,
        reader: bufio.NewReader(conn),
        prefix: '+',
    }, nil
}
```

### 7.2 Sending Commands & Parsing Responses

```go
type Response struct {
    CommandEcho string            // e.g. "set_freq: 14250000" or "get_freq:"
    Data        map[string]string // key-value pairs from data records
    Keys        []string          // ordered key list to preserve response order
    ReturnCode  int               // 0 = success, negative = error
}

func (c *Client) exec(cmd string) (*Response, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Send with ERP prefix
    _, err := fmt.Fprintf(c.conn, "%c%s\n", c.prefix, cmd)
    if err != nil {
        return nil, fmt.Errorf("write: %w", err)
    }

    resp := &Response{Data: make(map[string]string)}
    firstLine := true

    for {
        line, err := c.reader.ReadString('\n')
        if err != nil {
            return nil, fmt.Errorf("read: %w", err)
        }
        line = strings.TrimRight(line, "\r\n")

        // Rule 2: RPRT line ends the response block
        if strings.HasPrefix(line, "RPRT ") {
            code, _ := strconv.Atoi(strings.TrimPrefix(line, "RPRT "))
            resp.ReturnCode = code
            return resp, nil
        }

        // Rule 1: First line is the command echo
        if firstLine {
            resp.CommandEcho = line
            firstLine = false
            continue
        }

        // Rule 3: Data records are "Key: Value"
        if idx := strings.Index(line, ": "); idx > 0 {
            key := line[:idx]
            value := line[idx+2:]
            resp.Data[key] = value
            resp.Keys = append(resp.Keys, key)
        }
    }
}
```

### 7.3 Typed Wrapper Functions

```go
// GetFreq returns the current VFO frequency in Hz.
func (c *Client) GetFreq() (int64, error) {
    resp, err := c.exec("f")
    if err != nil {
        return 0, err
    }
    if resp.ReturnCode != 0 {
        return 0, &HamlibError{Code: resp.ReturnCode}
    }
    freq, err := strconv.ParseInt(resp.Data["Frequency"], 10, 64)
    return freq, err
}

// SetFreq sets the VFO frequency in Hz.
func (c *Client) SetFreq(hz int64) error {
    resp, err := c.exec(fmt.Sprintf("F %d", hz))
    if err != nil {
        return err
    }
    if resp.ReturnCode != 0 {
        return &HamlibError{Code: resp.ReturnCode}
    }
    return nil
}

// GetMode returns the current mode and passband width in Hz.
func (c *Client) GetMode() (mode string, passband int, err error) {
    resp, err := c.exec("m")
    if err != nil {
        return "", 0, err
    }
    if resp.ReturnCode != 0 {
        return "", 0, &HamlibError{Code: resp.ReturnCode}
    }
    mode = resp.Data["Mode"]
    passband, _ = strconv.Atoi(resp.Data["Passband"])
    return mode, passband, nil
}

// SetPTT sets PTT state: 0=RX, 1=TX, 2=TX Mic, 3=TX Data.
func (c *Client) SetPTT(ptt int) error {
    resp, err := c.exec(fmt.Sprintf("T %d", ptt))
    if err != nil {
        return err
    }
    if resp.ReturnCode != 0 {
        return &HamlibError{Code: resp.ReturnCode}
    }
    return nil
}
```

### 7.4 Error Type

```go
type HamlibError struct {
    Code int
}

var errNames = map[int]string{
    0:   "OK",
    -1:  "EINVAL",
    -2:  "ECONF",
    -3:  "ENOMEM",
    -4:  "ENIMPL",
    -5:  "ETIMEOUT",
    -6:  "EIO",
    -7:  "EINTERNAL",
    -8:  "EPROTO",
    -9:  "ERJCTED",
    -10: "ETRUNC",
    -11: "ENAVAIL",
    -12: "ENTARGET",
    -13: "BUSERROR",
    -14: "BUSBUSY",
    -15: "EARG",
    -16: "EVFO",
    -17: "EDOM",
}

func (e *HamlibError) Error() string {
    if name, ok := errNames[e.Code]; ok {
        return fmt.Sprintf("hamlib: %s (RPRT %d)", name, e.Code)
    }
    return fmt.Sprintf("hamlib: unknown error (RPRT %d)", e.Code)
}
```

### 7.5 Checking VFO Mode

```go
func (c *Client) CheckVFOMode() (bool, error) {
    resp, err := c.exec("\\chk_vfo")
    if err != nil {
        return false, err
    }
    // chk_vfo is special: in default protocol it returns "CHKVFO x"
    // In ERP it returns via data records. Handle both.
    if v, ok := resp.Data["VFO"]; ok {
        return v == "1", nil
    }
    return false, nil
}
```

### 7.6 Querying Supported Tokens

Many commands support a `?` query to discover what the backend supports:

```go
func (c *Client) SupportedModes() ([]string, error) {
    resp, err := c.exec("M ?")
    if err != nil {
        return nil, err
    }
    // The response will be in the data record
    // Note: the exact key may vary. Parse the first data value.
    for _, key := range resp.Keys {
        return strings.Fields(resp.Data[key]), nil
    }
    return nil, nil
}
```

### 7.7 Timeout & Reconnection

```go
// Set read/write deadlines for robustness
func (c *Client) exec(cmd string) (*Response, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.conn.SetDeadline(time.Now().Add(5 * time.Second))
    defer c.conn.SetDeadline(time.Time{}) // clear deadline

    // ... same send/receive logic as above
}
```

---

## 8. Wire Protocol Summary Diagram

```
CLIENT                                         SERVER (rigctld)
  │                                               │
  │  TCP connect to :4532                         │
  │──────────────────────────────────────────────►│
  │                                               │
  │  "+f\n"                  (get frequency, ERP) │
  │──────────────────────────────────────────────►│
  │                                               │
  │  "get_freq:\n"                (command echo)  │
  │◄──────────────────────────────────────────────│
  │  "Frequency: 14074000\n"       (data record)  │
  │◄──────────────────────────────────────────────│
  │  "RPRT 0\n"                    (end marker)   │
  │◄──────────────────────────────────────────────│
  │                                               │
  │  "+F 14250000\n"         (set frequency, ERP) │
  │──────────────────────────────────────────────►│
  │                                               │
  │  "set_freq: 14250000\n"       (command echo)  │
  │◄──────────────────────────────────────────────│
  │  "RPRT 0\n"                    (end marker)   │
  │◄──────────────────────────────────────────────│
  │                                               │
  │  "+M USB 2400\n"                    (set mode) │
  │──────────────────────────────────────────────►│
  │                                               │
  │  "set_mode: USB 2400\n"                       │
  │◄──────────────────────────────────────────────│
  │  "RPRT 0\n"                                   │
  │◄──────────────────────────────────────────────│
  │                                               │
  │  "+F bad_value\n"              (invalid input) │
  │──────────────────────────────────────────────►│
  │                                               │
  │  "set_freq: bad_value\n"                      │
  │◄──────────────────────────────────────────────│
  │  "RPRT -1\n"                   (EINVAL error)  │
  │◄──────────────────────────────────────────────│
```

---

## 9. References

- **rigctld man page:** https://hamlib.sourceforge.net/html/rigctld.1.html
- **rotctld man page:** https://hamlib.sourceforge.net/html/rotctld.1.html
- **ampctld man page:** https://hamlib.sourceforge.net/html/ampctld.1.html
- **Error codes in rig.h:** https://github.com/Hamlib/Hamlib/blob/master/include/hamlib/rig.h
- **Command parser source:** https://github.com/Hamlib/Hamlib/blob/master/tests/rigctl_parse.c
- **Hamlib Wiki:** https://github.com/Hamlib/Hamlib/wiki

## 10. Remarks

The critical insight is the + prefix. Every command you send should be prepended with + to
activate the Extended Response Protocol. This transforms the protocol from "guess how many
lines the response has" into a deterministic structure where you always read lines until you
hit RPRT x. Without ERP, parsing is fragile because different commands return different numbers
of value lines with no labels.

Serialization matters. Since the protocol is request/response over a single TCP stream with
no multiplexing or request IDs, you must serialize command execution with a mutex. Sending
two commands concurrently will interleave their responses and corrupt your parser state.

The response key names vary slightly between commands. For example, get_freq returns `Frequency:`
for rigs but `Frequency(Hz):` for amplifiers. Your parser should handle this gracefully — the Go
implementation stores keys exactly as received and lets typed wrapper functions know which
key to look up.

Short commands vs. long commands are functionally identical — `+f\n` and `+\get_freq\n` produce
the same response. Short commands are more compact for the wire; long commands are more readable
for debugging. Pick one style and stay consistent.

The document covers all three daemons (rig, rotator, amplifier), every command with its
arguments and response keys, all 17 error codes, and includes a ready-to-use Go client skeleton
with connection management, parsing, typed wrappers, and error handling.
