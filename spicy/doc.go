/*
The 'spicy' package parses and can verify spicy signatures
(both the witness signatures and the chain of hashes that
transfers trust from there through to the spicy-signed content).

Spicy signatures are not produced by this package:
see the sibling 'signing' package for that.
(These are separate because they have very different dependencies:
verifying a spicy signature is much simpler than creating one!)
*/
package spicy
