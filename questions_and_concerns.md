# Ongoing List of Questions / Concerns for later

- Do I need to deal with offsets somewhere - i.e. making sure I'm storing integers/characters at the proper offset in the block so it works? I know Java handles that automatically, but what about Go?
- When I print a list of bytes, why do integers always show up as double the amount I'm expecting? (e.g. a byte labeled 1 will print as 2)
- Am I still filling every other slot or did I fix that?
- How exactly do file permissions work and when I'm creating files (or the db directory) what permissions should they have? (Should that be hard-coded?) How/where do I make sure that files are actually stored inside the db folder?
- Are slices which are made with a given capacity and never removed actually the same efficiency as arrays? Why can't I use arrays without a lot of stuff breaking? (Is it possible to pass a constant number to a function and then return an array of that size?)
- Should I just change everything to int64?
- Can I really fill a whole I/O memory page and be fine? Is the metadata for it really separate? Or do I have to leave some room for the OS?
- Should all objects (structs) be defined on a single page, or is defining them in their respective files better?
- Is there some way to have setInt vs setString combine better, instead of duplicating each other so much?

## Parking Lot (good but not priority)

- Make a lot of structs contain anonymous other structs for easier reference without as many .whatevers
- Figure out where errors should come out in the process (especially reading/writing errors)
- Other fieldtypes?
- CONCURRENCY EVERYTHING (check out sync package)
- Change strategy for which unpinned buffer to choose to Clock variant
- test flag to control prints and stuff / better tests
- Choose which functions should be publicly available and capitalize them. Make sure everything _else_ is lowercase.
- Change buffer pool get-new-unpinned-buffer thing to use `wait` and wait for an unpin rather than doing it in a timed way.