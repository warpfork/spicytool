todo
====

(Not an exhaustive list!)

- [ ] the readme should link heavily to c2sp docs instead of redeclaring things here.
	- ...  but some of those must also first be written ;)
- [ ] code for parsing spicysig format should probably be hoisted to an upstream (torchwood?).


longer road
-----------

- The MVP of SpicyTool is creating a tlog locally using posix drivers of tessera -- supporting submission to remote tlogs and awaiting the return of a usable checkpoint from such a remote tlog is future work.
- The MVP of SpicyTool is not yet proactively reaching out to multiple witnesses when creating checkpoints.
- The MVP of SpicyTool operates blockingly: starting the writing of an entry, awaiting the witnesses to produce a fully checkpoint that includes their signatures, producing the MIP, and finally writing out the Spicy Sig file, all happen sequentially.  Creating large numbers of spicy sigs rapidly will probably want a batch mode of operation!
