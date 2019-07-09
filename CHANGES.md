CHANGES
=======

0.3.0 (2019-08-09)
------------------

NEW:

* catalog: Multiple repositories can now be registered and serve as recipes
  sources, thus supporting private repositories.
* cmd/mkdeb: A "lint" subcommand has been added to perform basic recipes
  linting.
* cmd/mkdeb: The "clean" subcommand has been renamed to "cleanup".
* cmd/mkdeb: A "--no-emoji" command flag can be specified to disable emoji
  printing.
* cmd/mkdeb: A folder can be specified as "--from" flag argument to use as local
  sources instead of an archive or a single file.
* cmd/mkdeb: Commands outputs has been refined to ensure better consistency.

0.2.0 (2019-04-22)
------------------

NEW:

* recipe: A "file" source type can be specified in recipes rules if source is
  not an archive.
* cmd/mkdeb: A "--recipe" command flag can be specified in build commands to
  specify a custom recipe path.
* recipe: An exclusion pattern can be added to recipes install rules to exclude
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
