# Dynamic DNS Daemon
Dead Simple Dynamic DNS Updater

On execution creates or updates the IPv4 A record for a specific domain. Errors if there is more than one existing A record for specified name.

## Usage
Set environment variables:
- `CLOUDFLARE_API_TOKEN` The Cloudflare API token (not API key)
- `DNS_A_RECORD_FQDN` The FQDN to use in the A record
- `CLOUDFLARE_SITE_ZONE_ID` The Zone ID for the site

Run the binary.
