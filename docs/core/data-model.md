# Core Data Model (minimal)

This document defines the minimal tables required for basic Discord identity data.

## Tables

### guilds

| column | type | description |
| --- | --- | --- |
| id | text | Discord Guild ID |
| created_at | timestamptz | created time (UTC) |
| updated_at | timestamptz | updated time (UTC) |

### users

| column | type | description |
| --- | --- | --- |
| id | text | Discord User ID |
| is_bot | boolean | bot flag |
| created_at | timestamptz | created time (UTC) |
| updated_at | timestamptz | updated time (UTC) |

### guild_members

| column | type | description |
| --- | --- | --- |
| guild_id | text | Discord Guild ID |
| user_id | text | Discord User ID |
| joined_at | timestamptz | join time (UTC) |
| left_at | timestamptz | leave time (UTC, nullable) |
| created_at | timestamptz | created time (UTC) |
| updated_at | timestamptz | updated time (UTC) |

## Constraints

- guilds.id: primary key
- users.id: primary key
- guild_members: primary key (guild_id, user_id)
- guild_members.guild_id references guilds.id
- guild_members.user_id references users.id
