# BACKLOG

- fix: when a glance file regenerates, *all* glance files in *all* of its parent directories need to regenerate. because a parent glance file makes use of the glance files in its subdirectories, they risk losing sync if they are not regenerated when their children are regenerated

- create post-commit hook to run glance
- improve performance -- make *fast*
- timestamp generated glance files
- remove force option
- audit whole codebase against dev philosophy, identify key things to hit
- refactor aggressively
