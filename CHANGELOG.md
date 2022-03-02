# Changelog
* Version 1.2.0
    - Add Sentry Error Logging

* Version 1.0.5
    - Fix syntax error in logrotate file. Now the server should get a USR1 signal to reopen log files
    - Fix seg fault in noid-tool utility. (Does not affect the server)

* Version 1.0.4
    - Log more errors
    - If configured for a database backend, at startup only continue if connected.

* Version 1.0.3
	- Move repository to github.com/ndlib
	- Update sqlite3 package
	- Log minted ids
	- Add end-to-end tests
	- Use application/json content type when appropriate
	- Don't log database passwords

* Version 1.0.2
	- Add /stats route. For now it only reports the version running.

* Version 1.0.1
	- Add storagedir config file option.

* Version 1.0.0
	- Initial Release
