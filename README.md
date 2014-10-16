moreio
======

Some additional IO utilities for Go.
 * *RollingReader*: Concatenate an arbitrary number of [`io.Reader`](http://golang.org/pkg/io/#Reader)s into a single Reader. Like [`io.MultiReader`](http://golang.org/pkg/io/#MultiReader), but supports addition of Readers during consumption. Thus, a RollingReader requires manual closure.
 * *SharedBuffer*: Buffer which supports multiple concurrent readers. Flushes the portion of the buffer which has been read by all.
