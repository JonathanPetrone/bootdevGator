# Gator the blog aggregator

## Prerequisites

This program requires the following to be installed on your system:
* Go, recommended version 1.23.4
* PostgreSQL (version 16.6 or later)

### Verifying Your Installation

To verify you have the required software installed:

1. Check Go installation:
   go version

2. Check PostgreSQL installation:
   psql --version

### Installation

Install the gator CLI by running:

```go install github.com/jonathanpetrone/bootdevGator@latest```

Make sure your `$GOPATH/bin` directory is in your PATH to run `gator` from anywhere.

### Configuration

1. Create a config file named `config.json` in the root directory with the following structure:
   {
     "database_url": "postgresql://username:password@localhost:5432/gator_db?sslmode=disable"
   }

   Replace username, password, and gator_db with your PostgreSQL credentials.

### Usage

gator login      // Login to your account
gator register   // Create a new account
gator reset      // Reset all users (admin only)
gator users      // List all users
gator agg        // Run the feed aggregator
gator addfeed    // Add a new feed URL (requires login)
gator feeds      // List all feeds
gator follow     // Follow a feed (requires login)
gator following  // List feeds you're following (requires login)
gator unfollow   // Unfollow a feed (requires login)
gator browse     // Browse your feed entries (requires login)
