# goannoy
Approximate Nearest Neighbors in golang optimized for memory usage and loading/saving to disk.

:warning: Thus uses lots of `unsafe` to handle union and the c++ style allocations. But loading a index and using it is then blazing fast since it just maps the data from the loaded memory.

:bulb: This is not yet complete, come back again within a month and then all distances should be ported and all bugs vetted.

## Credits

This is a port of the Spotify https://github.com/spotify/annoy
