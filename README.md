moreio
======

Some additional IO utilities for Go.
 * *RollingReader*: Concatenate an arbitrary number of [`io.Reader`](http://golang.org/pkg/io/#Reader)s into a single Reader. Like [`io.MultiReader`](http://golang.org/pkg/io/#MultiReader), but supports addition of Readers during consumption. Thus, a RollingReader requires manual closure.
 * *SharedBuffer*: Buffer which supports multiple concurrent readers. Flushes the portion of the buffer which has been read by all.
 * *Meters*: Wrappers for io.Readers and io.Writers which count total amount of bytes read and written, respectively.
 * *Stream*: Encoder and Decoder for a stream of undefined length. It uses a chunked transfer encoding, where each chunk's length is specified in front of the chunk.
