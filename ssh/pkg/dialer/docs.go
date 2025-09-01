// Package dialer provides utilities to manage and use reverse connections
// opened by agents so the server (or other services) can dial back into a
// device. The package supports two transport modes (protocol versions): the
// legacy revdial-based HTTP transport (v1) and a yamux-multiplexed transport
// (v2). When using v2, per-stream application protocols are negotiated using
// multistream identifiers defined in this package.
//
// # High level concepts
//
//   - Manager: a connection manager that stores active reverse transports and
//     exposes methods to bind new agent connections and to dial a device by
//     its key. It also exposes callbacks for tracking when connections are
//     closed or when keep-alive events occur.
//
//   - Dialer: a thin wrapper around a Manager which also holds an
//     internalclient to perform device lifecycle operations (heartbeat /
//     offline notifications) and provides DialTo which returns a ready-to-use
//     net.Conn for a requested Target.
//
//   - Target: an interface implemented by small helpers that prepare a raw
//     connection for a particular application-level purpose (for example,
//     opening or closing an SSH session, or establishing an HTTP proxy). The
//     prepare method will perform any necessary handshake depending on the
//     negotiated connection version.
//
// # Versioning
//
// ConnectionVersion1 (v1) uses the older revdial/http handshake where the
// client expects HTTP-style GET/CONNECT requests. ConnectionVersion2 (v2)
// uses a yamux session and performs per-stream negotiation with the
// multistream protocol strings (see ProtoSSHOpen, ProtoSSHClose,
// ProtoHTTPProxy). Callers should prepare the appropriate Target and the
// dialer will perform the correct handshake based on the returned
// ConnectionVersion.
//
// Usage (server-side)
//
// Typical server usage is:
//   - When an agent connects, call Manager.Bind(tenant, uid, conn) to
//     register the reverse transport. The manager will keep the session alive
//     and call configured callbacks on events.
//   - To connect to a device, create a Dialer (NewDialer) and call
//     Dialer.DialTo(ctx, tenant, uid, target). DialTo returns a net.Conn
//     already prepared for the requested target (or a raw connection if the
//     target is nil).
//
// The package intentionally keeps the wire-level protocol identifiers and
// version handling colocated with the dial logic so the agent and server
// implementations can remain compatible and easy to reason about.
package dialer
