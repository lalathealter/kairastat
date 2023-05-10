# About the API

Please note that in order to access an API you should use a path of `./api` and not a root one; 

## POST ./api

Creates an event of the given name with the client's associated IP address; If an event with that name for that IP address already exists, increases an event counter; To pass arguments use URL queries like this:
`POST ./api?event={event_name_value}&authorized`

1. the `event` parameter:
    - is required (must be passed) 
    - represents a user event's name
    - accepts strings no longer than 128 characters
2. the `authorized` parameter:
    - is optional
    - can hold any value; If the parameter is present, it's considered to be a boolean true equivalent;
    - updates a user's status even when absent (in that case, sets the authorization status to false)
    - updates a user authorization status so their events could be filtered by their current status 

## GET ./api

Retrieves the data about events based on provided filters and aggregators; To provide those, use URL queries like this:
`GET ./api?{filter-option}={filter_argument}&{aggregator-option}={aggregator_argument}`
Please note that filter and aggregator options are exclusive; If you provide several filter (or aggregator) options at once, a service will respond with only one of them (depending on what would be checked first in the corresponding handler function)
If no aggregator is passed, API returns all available events;

Available filters: 
    1. `starts-with`:
        - filters events based on a provided name start
        - utilizes SQL's LIKE statement, so you would get an expression of `{your_argument}%`
    2. `later-than`:
        - filters events based on the time they were originally created
        - accepts only valid RFC3339 time strings like `2006-01-02T15:04:05+07:00` (the plus sign indicates a start of a timezone and can be replaced with a minus)

Available aggregators:
    1. `event`:
        - aggregates exact name matches
    2. `user-ip`:
        - aggregates events by a provided IP address
        - must be a string representation of a **IPv4 convertable address**
    3. `is-authorized`:
        - aggregates events based on it's author's current authorization status
        - parameters of "true", "t", "1", "yes" and "y" are equivalent to a boolean true value (the character case doesn't matter)
        - anything else is considered a boolean false value 

