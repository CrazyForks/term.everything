# 0.7.8
- Added support for mouse when really zoomed out by
    - Adding SGR 1006 support for mouse reporting in terminals that support it.
# 0.7.7
- Fixed a bug where the terminal would not output SIXELS even when the override was set.
# 0.7.6
- Now it only redraws the screen when there are changes (like frame draw requests from wayland, mouse movement, etc), reducing CPU usage, and network bandwidth for remote sessions.
- Moved profiling code behind a build tag to reduce binary size for normal users.
# 0.7.5
- Converted code to entriely go (with a tiny bit of c). THis simplifies the build process and reduces dependencies. This lets us easily port to new platforms like arm64.

# 0.5.4
- Added environment variables to pass through chafa options.
- Added environment variable `TERM_EVERYTHING_PIXEL_TYPE` to workaround certain cases where there is a bgr/rgb swap.
- Added `--max-frame-rate` option to ...set the maximum frame rate.
- Fixed a build bug with node_canvas's included version of libstdc++.so
# 0.5.3
fixed right mouse click and sub menus
Set XDG_SESSION_TYPE to wayland so app that have
a wayland support will use it.
# 0.5.2
fixed scrolling inversion https://github.com/mmulet/term.everything/issues/4
# 0.5.0
First Beta release
