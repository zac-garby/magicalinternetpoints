# magicalinternetpoints

A way to reward and acknowledge all that time you've spent doing
nothing of use on the internet!

## Running the server

You'll need

 - A database, `magicalinternetpoints.sqlite3`
   - You can make this from `schema.sql` with the command `sqlite3 magicalinternetpoints.sqlite3 < schema.sql`.
   - It'll be empty, but for any functionality you'll want to add some sites and point sources.
 - A GitHub OAuth app, and a Reddit OAuth app
 - A `.env` file, exporting environment variables:
   - `MIP_PORT`, the port to run the server on
   - `MIP_BASEURL`, the base URL on which the server will run
   - `GITHUB_TOKEN`, the GitHub OAuth ID
   - `GITHUB_SECRET`, the GitHub OAuth secret
   - `REDDIT_TOKEN`, the Reddit OAuth ID
   - `REDDIT_SECRET`, the Reddit OAuth secret

Then, run `source .env && go run .`