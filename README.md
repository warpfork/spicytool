SpicyTool
=========

`spicytool` is a CLI tool to generate Spicy Signatures.
You can use it in roles similar to how you'd use `gpg -s`.

Spicy Signatures provide authenticity guarantees in ways similar to familiar signature systems based on asymmetric cryptography primitives,
but are based on append-only verifiable log structures instead of asymmetric crypto.
These "transparency logs" allow both verifiable sequencing of events, and aggregation of witnessing of events,
in ways that offer security even beyond what simple asymmetric signatures can provide in isolation.

`spicytool` has two main functions:

- `spicytool verify` takes a Spicy Signature file, verifies the signatures on it, and verifies that the merkle proof transfers those signatures to apply to the Spicy Signature's subject.
- `spicytool sign` takes an arbitrary file or text stream, computes its hash, appends it to a transparency log (which is locally maintained, using Tessera with its posix filesystem backend), and generates the Spicy Signature file.

Because spicytool is maintaining a transparency log itself, it also necessarily has some commands for administrating that:

- `spicytool keygen` has will generate keys which are suitable to use for operating log and having it sign its own checkpoints.

The sign operation does include reaching out to all of the log's witnesses, proactively, and asking them to sign a fresh checkpoint of the log.
As a result of this, there is some network latency involved in the sign operation.

SpicyTool is a work in progress.

Visit https://c2sp.org/ for more information about Spicy Signatures and Transparency Logs,
and the primitives and implementation design choices behind them!


License
-------

MIT
