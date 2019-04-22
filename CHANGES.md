CHANGES
=======

0.2.0 (2019-04-22)
------------------

NEW:

* recipe: A "file" source type can be specified in recipes rules if source is
  not an archive.
* cmd/mkdeb: A "--recipe" command flag can be specified in build commands to
  specify a custom recipe path.
* recipe: A exclusion pattern can be added to recipes install rules to exclude
  files from resulting packages.
* archive: Non compressed tar archives are now supported.

BUG FIXES:

* archive: Fix mishandled files mode reading from source archives.
* deb: Fix missing trailing slashes for directories references in generated
  packages.
* cmd/mkdeb: Fix a crash when search results return partial data.
* cmd/mkdeb: Fix a crash when source format is unsupported.
* cmd/mkdeb: Fix download progress issue when Content-Length header is missing.

0.1.0 (2018-05-12)
------------------

* Initial release
