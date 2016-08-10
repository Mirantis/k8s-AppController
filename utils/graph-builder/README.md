# Running
To run, just open `index.html` in directory of your local copy using `gnome-open index.html` under unixoids, `open index.html` under macs, or just clicking it in file browser of your system/choice.

You can also serve this directory from http, e.g. `python -m SimpleHTTPServer 9000` and load `0.0.0.0:9000` in your browser.

# Usage
Right click on a node to remove it, or move it's name to either parent or child input (convenient when you are adding edges to graph).

Use forms below the graph to add new nodes and edges. Removing edges is not supported yet in the graph itself.

Use `generate` button to dump your current graph to dependencies yaml.

Use `load` button to load your yaml from textarea into the graph.
