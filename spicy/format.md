The Spicy Signature Format
==========================

(This is a provisional draft: a more formal specification can be
expected to appear later at https://c2sp.org/ .)

Spicy Signatures are a textual format containing
an offline-verifiable proof that some data has been logged to a transparency log
and that its inclusion has been observed by some number of witnesses.

More specifically, a Spicy Signature composes the following information
which is derived from an entry and its presence in a transparency log:

- a checkpoint of a tree state of that log,
- a collection of witness signatures over that checkpoint,
- and a MIP (Merkle Inclusion Proof) of some entry's presence in the log.

See the "Considerations" section, below,
for more information about ways to apply Spicy Signatures in various situations.


Format
------

Roughly:

```
c2sp.org/spicy-signature@v1
index {entryIndex:b10int}
{MIPhash1:b64string}
{MIPhash2:b64string}
{MIPhashN:b64string}

{logname:urlishString}
{treeHeight:b10int}
{treeRootHash:b64string}

— {logname:urlishString} {witnessSignature0:b64string}
— {witness1:urlishString} {witnessSignature1:b64string}
— {witness2:urlishString} {witnessSignature2:b64string}
— {witness3:urlishString} {witnessSignature3:b64string}

contexthint
{arbitrary:string}
```

A slightly more literal example (although still with elisions for human reading) would be:

```
c2sp.org/spicy-signature@v1
index 73894
gSKyXoYZUgZ6jduWYrkDOARinOMGJveXjgMkBTcdPlQ=
B95lDa8R83lS8n0eG+o0buTxRKQTYFi//1U8anccXmA=
EKNzoDWG8LGC0Yp9o+sv3qllpMP9uHQ9B20KNL+Q1zs=
RoopEkOdqkYqMB4MJXrbt/hMjOxsVn0IrWjpz1ZMMes=
AHCioX9nLjsrse6YhjRRmk1WUEirVOLLRoOQ6vfO5vk=

example.com/fancylog
109482
sFodV/vSp5O8n9a8QpW6PRY97tfOSW5bsc2Xl/EQi08=

— example.com/fancylog hI2DJw[...]1roloI=
— witness1.example mJirIklj[...]qY9v2B/5bg==
— witness2.example TnKKVHLX[...]xwYwrSjgow==
— witness3.example S4X82uH5[...]3oEcROGLFQ==

contexthint
age-v1.2.1-darwin-arm64.tar.gz
```

### a spicy signature begins with an entry index and a MIP

A spicy signature states an index number, as a base 10 integer, that is the index into a transparency log of an entry that the signature applies to.

(The value itself that the signature applies to is not serialized in a spicy signature!
When verifying a spicy signature, that value is expected to be recomputed from the data in hand.
This is to ensure that the verification is applying to real data, rather than vaccuously applying to itself.)

Following the index number is a series of lines which are the base64 encodings of entries in a Merkle Inclusion Proof.

A MIP is also known as a Merkle Audit Path.
A full specification of this format and its computation is found in https://tools.ietf.org/html/rfc6962#section-2.1.1 .

Due to the way this MIP structure elides all nodes that are regenerable during a successful verification operation,
both the index (included in this section) and the tree height (included in a following section) are necessary pieces of information,
because they indicate which side of each branching in the Merkle tree these MIP entries occupy.

The log that the entry index refers to is named later, in the signed note section which contains the checkpoint.
However, this name is not generally load-bearing.
(While the index number is defined in the context of that log, verification only uses that number in a mathematical way;
it does not ever prompt an actual need to read any information from the log that's not already contained in the Spicy Sig!)

### a spicy signature contains a signed note that is a checkpoint of a transparency log

The middle section of a Spicy Signature describes a state of the transparency log.

This information contains a name of the log,
its tree height at the point in time that the Spicy Signature was generated (and is the position which the MIP is computed against),
and the root hash of the tree at that time.

This information is wrapped in a Signed Note format,
which contains a series of named signatures.

- Checkpoints are detailed here: https://c2sp.org/tlog-checkpoint
- Signed Notes, more generally, are detailed here: https://c2sp.org/signed-note

### the actual entry in the log is a combination of content (not in-band) and optionally the context hint

The last section of a spicy sig file is an optional "context hint" section.

This may be easiest to explain by examining a concrete application:
when we use Spicy Signatures to sign files with SpicyTool,
then the file body is the "content", and the contexthint is set to the filename seen at the time of signing.

While content is *not* distributed with a spicy signature, the context hint (often) *is* --
this allows the signature to be verified on the content, and does not require the content hint to be transferred out of band.

In application: this means SpicyTool can verify a signed file,
and while it may attempt to derive the context hint at verify time in the same way as it did at sign time (by looking at the filename!),
it can also still verify the signature even if the filename is not the same as the one the signature was created for.
(The tool emits still halts in case of such a mismatch, but it can distinguish this from an opaque verification failure.)

A context hint is freeform text.
The entire trailer of the file subsequent to the line containing the "contexthint" keyword is considered to be the context hint.

TODO: specify and document the exact munge that compiles the content and context hint into an entry.

### errata

Parser implementors may wish to note that the number of entries in the MIP section may be zero.
This is rare, but true for the very first entry in a log!

In the examples above, the reader may note that the log name is seen appearing twice in the same document:
this is not a typo;
it is because a log signs itself as a witness of itself.
However, that signature should not be considered privileged or special compared to other witness signatures.
It is also not required; it is just common to include
because it is operationally easy to do so.



Considerations
--------------

### Considerations for Usage

An expected use of Spicy Signatures is to use their publication to
make it publicly undeniable that the signature was issued.
Distributing a Spicy Signature along with some data is a way to
provide a consumer of that data with confidence that the data
they are given is the same data everyone else is being given...
and if that is untrue, they will at least be able to prove that.

Spicy Signatures are a general transparency system primitive.
They are not specific to a realm of application (e.g. Key Transparency, Certificate Transparency, Binary Transparency, etc).
Applying Spicy Signatures to specific domains generally involves
making a structured entry format which suits the domain,
and applying the spicy signature over that.

The entries that a Spicy Signature attests to may have any form,
so long as that entry is also accepted by the underlying transparency log.
(Generally this means entries must be of limited size; and typically,
hashes of content are used as a means to describe arbitrary data within a limited entry size.)
In many cases, it is sufficient to regenerate those entries from
data as it is already distributed in that domain (and so no additional
storage is required to attain useful properties);
in other cases, unlocking value in a domain may require more than
only undeniable claim issuance (for example, it may also require
providing additional related information that a monitoring system requires),
and that may require an additional out-of-band mechanism for distributing
that additional information.
