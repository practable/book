# book
This package implements bookings with arbitrary durations

((This package is under active development and is not yet ready for use))

## Features

- Advance bookings 
- Arbitrary intervals for bookings
- Guarantee exclusive bookings
- Different UIs for different users
- Different limits on number of current/future bookings for different users
- Different usage limits for different users
- Resource status tracking (e.g. if offline for maintenance or failing tests)

## Limitations
- Redundancy is not supported - it's doable, but too complicated to reason about hiearahical groupings of kit at this stage where some people book a type of experiment and others book specific examples.
- Unlimited access to simulations (unlimited, anytime, no need for booking?) - relative simple, but out of scope for now (simple mod for later).
- What if kit doesn't work - how to get another? If cannot cancel ...?! For now - when taking up booking, it returns, equipment not available. More advanced options can be added later (e.g. admin edits bookings to substitute another piece of kit instead, when taking an experiment offline).

## Cancel after Start implemention thoughts

include booking names in the stream tokens 
update the relay API to reflect new stream token field (booking_name) AND the deny endpoint
generate a client from the swagger
use that client to submit the stream token to the deny list
For testing interval, run a relay instance locally

if cancellation is after start, then submit token to the relay deny list. IF successful, then cancel the booking. If not, error, suggest try again? That way, do not cancel something that is not actually cancelled at the relay.


## Usage suggestions

Use same booking page link format as at present, except when using this booking system, GET instead of a token, a file with a list of policies that apply to that link (e.g. everyone would be a list of the policies for all the different experiments, that don't conflict with class usages, whereas an superadmin would have all policies, including those for exclusive use of classes)

In client side javascript, prepare calendar info e.g. .ics for their student outlook account, so students can take of reminding themselves automatically, but without us having to store identifying info like emails


### Implementation limitations
- booking status e.g. cancelled, unfulfilled, started are ignored in the Replace(Old)Bookings methods
- we don't record the time an activity is generated for a booking (could find this in the server logs perhaps), i.e. we can't collect stats on when people arrive for their bookings.
## Implementation overview

### Definitions

A `resource` represents a single bookable entity which is a real, physical piece of equipment. The entity can be booked for any arbitrary interval, so long as it does not overlap with any other booking. Booking exclusivity is guaranteed at the resource level. A resource has an availability (bool) and a status message (e.g. "Test failed on (date)", or "Test passed on (date)")

A `filter` represents a collection of `allowed` and `denied` periods, allowing for the implementation of policies about when a booking can be made.

A `slot` represents access to one resource, where the access is restricted by a `filter`. There are potentially many slots for one resource, and slots may overlap, resulted in shared access (but the first to book a resource gets exclusive use of it, it's just the opportunity to book it that is shared). The slot lists the user interface set that can be used with the hardware.

A `user` represents an entity that can book `slots` according to zero or more `policy` instances.

A `policy` represents the maximum usage permitted for a list of one or more slots, the minimum or maximum length of bookable interval, the maximum number of current/future bookings.

#### What happened to pools?

Now that we are booking in advance, the concept of redundancy as it applies to different hierarchies of groupings of equipment is complex, and so pools and redundancy is left for future work. Instead for now, if a user books something, and it becomes unavailable, then they can rebook on something else. It may be upsetting to a user to have a session cancelled due to kit failure, and having to sort a replacement session themselves, but better we reduce frustration overall for many users by offering some form of booking.

#### Why is the filter in the slot, not the policy?

Some resources may be shared between two courses, other between three or four courses, and some exclusive users may only need a subset of the equipment of that type, so blanket filtering of access to a set of resources is not sufficient. Slots with individually-defined access times to specific resources can be added to a policy to express this, for example half the slots for the other class will be available any time, the other half only when not being used by the exclusive user. In other words, the slots in a policy may have different filters in them.  This does require editing all slots related to a resource to establish a new exclusive user, but that can be avoided later by using a rule-based system inspired by or based on Ory Keto / Google Zanzibar.

### Example Process Flow 

A course organiser and practable administrator agree a `policy` for use by the students in a class, including the times of access to equipment which is set by the slots assigned to the policy. Optionally additional policies for course staff and tutors/demonstrators could be agreed, e.g. with access to other user interfaces.

The policy sets user-specific aspects such as total usage, booking duration and live booking count. The slots contain any limitatations on when the equipment is available to a class. Thus, the policy, once agreed need not be changed, however, the definition of the slots it refers to, may change e.g. if another class needs the equipment exclusively for an hour, then that hour would be denied in the slots assigned to other classes. Such changes would happen behind the scenes, and not be explicitly shown in the policy.

A policy exists as a JWT token on a server. A user is provided with a link to a booking system with the policy-code as a URL param, as we currently do for setting groups.

A booking agent (client-side javascript, usually) checks to see if it is has already used the service before, by looking in its cache for pseudo-id. If no pseudo-id is found, it asks the user for their exam code, and checks/validates the format of the exam-code.

The user will see any bookings they have made under any other other policy, but can only book under the policy they currently have - they have to go to the link for the other course to book for that course. In this way, we can refresh the policy as needed, rather than having old policy codes hanging around or storing them in the booking system?  Simpler to test if we don't remember codes? Or store policy codes with expiration date? Get the policy fresh everytime anyway .... What if a code is shared widely? We can just change the link. So easier if we don't save old policies, because then we do not have to tell the booking system to invalidate old policies stored with users - unless we can auto-remove them if they are found to be expired?  These complexities can be better handled when we have the Ory version up and running, so don't get bogged down in it for now - just go with the live-only/non-stored policies for simplicity.

The policy is submitted along with any requests for availability or bookings, as the authorisation.

If a user has spare bookings available then can make a new booking. They can see availability, before cancelling their current booking(s), to make new ones. Don't offer a booking swap at this stage?

Cancel bookings is an option.

Store booking against the pseudo-id so it can be retrieved on future visits.

When a booking becomes due, submit the booking to the server for streaming stuff, same as before.

In this first version, don't allow session cancellation / i.e. can't cancel booking after launching it, to save having to re-do the user interfaces at this stage.

## A example manifest

This is the test manifest, but it has test strings for many fields which do not have the same appearance as the real world data so may be confusing as it is.
TODO edit strings to represent a real world use case

```
descriptions:
  d-p-a:
    name: policy-a
    type: policy
    short: a
  d-p-b:
    name: policy-b
    type: policy
    short: b
  d-r-a:
    name: resource-a
    type: resource
    short: a
  d-r-b:
    name: resource-b
    type: resource
    short: b
  d-sl-a:
    name: slot-a
    type: slot
    short: a
  d-sl-b:
    name: slot-b
    type: slot
    short: b
  d-ui-a:
    name: ui-a
    type: ui
    short: a
  d-ui-b:
    name: ui-b
    type: ui
    short: b
policies:
  p-a:
    book_ahead: 0s
    description: d-p-a
    enforce_book_ahead: false
    enforce_max_bookings: false
    enforce_max_duration: false
    enforce_min_duration: false
    enforce_max_usage: false
    max_bookings: 0
    max_duration: 0s
    min_duration: 0s
    max_usage: 0s
    slots:
    - sl-a
  p-b:
    book_ahead: 2h0m0s
    description: d-p-b
    enforce_book_ahead: true
    enforce_max_bookings: true
    enforce_max_duration: true
    enforce_min_duration: true
    enforce_max_usage: true
    max_bookings: 2
    max_duration: 10m0s
    min_duration: 5m0s
    max_usage: 30m0s
    slots:
    - sl-b
resources:
  r-a:
    description: d-r-a
    streams:
    - st-a
    - st-b
    topic_stub: aaaa00
  r-b:
    description: d-r-b
    streams:
    - st-a
    - st-b
    topic_stub: bbbb00
slots:
  sl-a:
    description: d-sl-a
    policy: p-a
    resource: r-a
    ui_set: us-a
    window: w-a
  sl-b:
    description: d-sl-b
    policy: p-b
    resource: r-b
    ui_set: us-b
    window: w-b
streams:
  st-a:
    audience: a
    ct: a
    for: a
    scopes:
    - r
    - w
    topic: a
    url: a
  st-b:
    audience: b
    ct: b
    for: b
    scopes:
    - r
    - w
    topic: b
    url: b
uis:
  ui-a:
    description: d-ui-a
    url: a
    streams_required:
    - st-a
    - st-b
  ui-b:
    description: d-ui-b
    url: b
    streams_required:
    - st-a
    - st-b
ui_sets:
  us-a:
    uis:
    - ui-a
  us-b:
    uis:
    - ui-a
    - ui-b
windows:
  w-a:
    allowed:
    - start: 2022-11-04T00:00:00Z
      end: 2022-11-06T00:00:00Z
    denied: []
  w-b:
    allowed:
    - start: 2022-11-04T00:00:00Z
      end: 2022-11-06T00:00:00Z
    denied: []
```

```
'descriptions:
  d-p-a:
    name: policy-a
    type: policy
    short: a
  d-r-a:
    name: resource-a
    type: resource
    short: a
  d-sl-a:
    name: slot-a
    type: slot
    short: a
  d-ui-a:
    name: ui-a
    type: ui
    short: a
policies:
  p-a:
    book_ahead: 0s
    description: d-p-a
    enforce_book_ahead: false
    enforce_max_bookings: false
    enforce_max_duration: false
    enforce_min_duration: false
    enforce_max_usage: false
    max_bookings: 0
    max_duration: 0s
    min_duration: 0s
    max_usage: 0s
    slots:
    - sl-a
resources:
  r-a:
    description: d-r-a
    streams:
    - st-a
    - st-b
    topic_stub: aaaa00
slots:
  sl-a:
    description: d-sl-a
    policy: p-a
    resource: r-a
    ui_set: us-a
    window: w-a
streams:
  st-a:
    audience: a
    ct: a
    for: a
    scopes:
    - r
    - w
    topic: a
    url: a
uis:
  ui-a:
    description: d-ui-a
    url: a
    streams_required:
    - st-a
ui_sets:
  us-a:
    uis:
    - ui-a
windows:
  w-a:
    allowed:
    - start: 2022-11-04T00:00:00Z
      end: 2022-11-06T00:00:00Z
    denied: []'
```



Intervals are specified as

```
type Interval struct {
	Start time.Time
	End time.Time
}

```


## Methods

### Maintenance commands

```golang
// PruneDiaries is a maintenance operation to prune old bookings from diaries
// to make booking decisions faster. There is an overhead to pruning trees
// because they are rebalanced, so don't do too frequently.
func (s *Store) PruneDiaries()

// PruneDiaries is maintenance operation that moves expired bookings from
// the map of current bookings to the map of old bookings
func (s *Store) PruneBookings() 

// PruneUserBookings is a maintenace operation to move
// expired bookings from the map of bookings but only
// to do so for a given user (e.g. ahead of checking
// their policy limits on future bookings).
func (s *Store) PruneUserBookings(user string
```

### User Commands

```golang
// GetDescription returns a description if found
func (s *Store) GetDescription(name string) (Description, error)

// GetPolicy returns a policy if found
func (s *Store) GetPolicy(name string) (Policy, error) 

// GetAvailability returns a slice of intervals for which a given slot is available under a given policy, or an error if the slot or policy is not found. The policy contains aspects such as look-ahead which may limit the window of availability.
func (s *Store) GetAvailability(policy, slot string) ([]interval.Interval, error)

// MakeBooking makes bookings for users, according to the policy
// If a user does not exist, one is created.
// APIs for users should call this version
func (s *Store) MakeBooking(policy, slot, user string, when interval.Interval) (Booking, error)

// CancelBooking cancels a booking or returns an error if not found
func (s *Store) CancelBooking(booking Booking) error

// GetBookingsFor returns a slice of all the current bookings for the given user
func (s *Store) GetBookingsFor(user string) ([]Booking, error) 

// GetActivity returns an activity associated with a booking, or an error
// if the booking is invalid in some way 
func (s *Store) GetActivity(booking Booking) (Activity, error) 

// GetOldBookingsFor returns a slice of all the old bookings for the given user
func (s *Store) GetOldBookingsFor(user string) ([]Booking, error)

// GetPolicyStatusFor returns usage, and counts of current and old bookings
func (s *Store) GetPolicyStatusFor(user, policy string) (PolicyStatus, error)

// GetPoliciesFor returns a list of policies that a user has booked with
func (s *Store) GetPoliciesFor(user string) ([]string, error)
```

### Admin commands

```golang
// GetSlotIsAvailable checks the underlying resource's availability
func (s *Store) GetSlotIsAvailable(slot string) (bool, string, error)

// SetSlotIsAvailable sets the underlying resource's availability
func (s *Store) SetSlotIsAvailable(slot string, available bool, reason string) error

// GetSlotBookings gets bookings as far as ahead as the policy will let you book ahead
// It's up to the consumer to handle any pagination
func (s *Store) GetSlotBookings(slot string) ([]diary.Booking, error)

// MakeBookingWithID makes bookings for users, according to the policy
// If a user does not exist, one is created.
// The booking ID is set by the caller, so that bookings can be edited/replaced
// This version should only be called by Admin users
func (s *Store) MakeBookingWithName(policy, slot, user string, when interval.Interval, name string) (Booking, error)

// ReplaceManifest overwrites the existing manifest with a new one i.e. does not retain existing elements from any previous manifests
// but it does retain non-Manifest elements such as bookings.
func (s *Store) ReplaceManifest(m Manifest) error

// CheckManifest checks for internal consistency, throwing an error
// if there are any unresolved references by name
func CheckManifest(m Manifest) (error, []string)

// ExportBookings returns a map of all current/future bookings
func (s *Store) ExportBookings() map[string]Booking

// ReplaceBookings will replace all bookings with a new set
// each booking must be valid for the manifest, i.e. all
// references to other entities must be valid.
// Note that the manifest should be set first
func (s *Store) ReplaceBookings(bm map[string]Booking) (error, []string)

// ExportBookings returns a map of all old bookings
func (s *Store) ExportOldBookings() map[string]Booking

// ReplaceOldBookings will replace the map of old bookings with the supplied list or return an error if the bookings have issues
func (s *Store) ReplaceOldBookings(bm map[string]Booking) (error, []string)

// ExportUsers returns a map of users, listing the names of bookings, old bookings, policies and
// their usage to date by policy name
func (s *Store) ExportUsers() map[string]UserExternal

// There is no ReplaceUsers (not implemented, not planned to implement, because  ReplaceOldBookings deletes and recreates the user list)

```

## API

[swagger](https://github.com/go-swagger/go-swagger)

Go to the root of the repo and issue:

```
$ swagger generate server -t serve -f ./api/booking.yml --exclude-main -A serve
$ go mod tidy
```
If making large changes, then there will be stale files left, so first
```
cd ./serve
rm -rf models
rm -rf restapi
```

### Parameters to set by environment variables

```
USER_NAME_MINIMUM_LENGTH -> getAccessTokenHandler
BOOKING_DURATION -> getAccessTokenHAndler
```

### Dev notes

Definitions

'#/definitions/Activity'
'#/definitions/Bookings'
'#/definitions/Description'
'#/definitions/Interval'
'#/definitions/Intervals'
'#/definitions/Manifest'
'#/definitions/Policy'
'#/definitions/Policies'
'#/definitions/PolicyStatus'
'#/definitions/Resource'
'#/definitions/Slot'
'#/definitions/SlotStatus'
'#/definitions/StoreStatusAdmin'
'#/definitions/StoreStatusUser'
'#/definitions/Stream'
'#/definitions/UI'
'#/definitions/UISet'
'#/definitions/User'
'#/definitions/Users'
'#/definitions/Window'


```yaml
  paginationExample:
    title: Slots 
    description: An array of slots representing available times to book a pool 
    type: object
    properties:
      available:
        description: how many activities are available in this pool
        type: number
        format: int64
      slots:
        description: an array of slots
        type: array
        items:
          $ref: '#/definitions/slot'
      is_truncated:
        description: whether the array of slots is truncated
        type: boolean
      max_slots:
        description: maximum number of slots returned per page
        type: number
        format: int64
      pool_id:
        description: pool_id so that can request next page using information in this object
        type: string
    required:
      - available
      - slots
      - is_truncated
      - max_slots
      - pool_id

```

### Activities format

#### Example of the activity information supplied by the original booking app
```
{"activities":[{"config":{"url":""},"description":{"further":"https://static.practable.io/info/penduino-real-1.0/index.html","id":"541975de-a47a-46cc-a140-c42798ec47f3","image":"https://assets.practable.io/images/booking/activities/penduino-real-1.0/image.png","long":"A single pendulum with electromagnetic drive system producing simple harmonic motion. The drive and braking are variable, with the coil energised over a specified angular range which can be increased or decreased to control the effect. The pendulum can also be slowed down by short-circuiting the coil, without applying any power, or left to swing freely with no drive or braking.","name":"Penduino","short":"A pendulum driven by an electromagnet, producing simple harmonic motion.","thumb":"https://assets.practable.io/images/booking/activities/penduino-real-1.0/thumb.png","type":"penduino-activity-v1.0"},"exp":1670703344,"streams":[{"for":"data","permission":{"audience":"https://relay-access.practable.io","connection_type":"session","scopes":["read","write"],"topic":"pend13-data"},"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b3BpYyI6InBlbmQxMy1kYXRhIiwicHJlZml4Ijoic2Vzc2lvbiIsInNjb3BlcyI6WyJyZWFkIiwid3JpdGUiXSwiYXVkIjpbImh0dHBzOi8vcmVsYXktYWNjZXNzLnByYWN0YWJsZS5pbyJdLCJleHAiOjE2NzA3MDMzNDQsIm5iZiI6MTY3MDcwMzA0NCwiaWF0IjoxNjcwNzAzMDQ0fQ.6bbZoNK-dCrRP_I-m6D1KQmOZ0_uHOaJlBnaAcI5trI","url":"https://relay-access.practable.io/session/pend13-data","verb":"POST"},{"for":"video","permission":{"audience":"https://relay-access.practable.io","connection_type":"session","scopes":["read"],"topic":"pend13-video"},"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b3BpYyI6InBlbmQxMy12aWRlbyIsInByZWZpeCI6InNlc3Npb24iLCJzY29wZXMiOlsicmVhZCJdLCJhdWQiOlsiaHR0cHM6Ly9yZWxheS1hY2Nlc3MucHJhY3RhYmxlLmlvIl0sImV4cCI6MTY3MDcwMzM0NCwibmJmIjoxNjcwNzAzMDQ0LCJpYXQiOjE2NzA3MDMwNDR9.NUvssbuUHIJNJ82AzExff2Rx3ckN2jH1lTQLAu_DsEs","url":"https://relay-access.practable.io/session/pend13-video","verb":"POST"}],"uis":[{"description":{"further":"https://static.practable.io/info/penduino-basic-ui-1.0/index.html","image":"https://assets.practable.io/images/booking/ui/penduino-default-1.0/image.png","long":"Start the pendulum swinging, and observe the difference in decay time for four different types of swinging - driving, braking, short circuit load, and open circuit coil.","name":"Penduino (Default)","short":"Control an Electromagnetic Pendulum","thumb":"https://assets.practable.io/images/booking/ui/penduino-default-1.0/thumb.png","type":"penduino-default-ui-1.0"},"streamsRequired":["data","video"],"url":"https://static.practable.io/ui/penduino-1.0/?streams={{streams}}&exp={{exp}}"}]}],"max":2,"msg":"For support contact timothy.drysdale@ed.ac.uk"}
```

Pretty-printed, that becomes:
```
{
   "activities":[
      {
         "config":{
            "url":""
         },
         "description":{
            "further":"https://static.practable.io/info/penduino-real-1.0/index.html",
            "id":"541975de-a47a-46cc-a140-c42798ec47f3",
            "image":"https://assets.practable.io/images/booking/activities/penduino-real-1.0/image.png",
            "long":"A single pendulum with electromagnetic drive system producing simple harmonic motion. The drive and braking are variable, with the coil energised over a specified angular range which can be increased or decreased to control the effect. The pendulum can also be slowed down by short-circuiting the coil, without applying any power, or left to swing freely with no drive or braking.",
            "name":"Penduino",
            "short":"A pendulum driven by an electromagnet, producing simple harmonic motion.",
            "thumb":"https://assets.practable.io/images/booking/activities/penduino-real-1.0/thumb.png",
            "type":"penduino-activity-v1.0"
         },
         "exp":1670703344,
         "streams":[
            {
               "for":"data",
               "permission":{
                  "audience":"https://relay-access.practable.io",
                  "connection_type":"session",
                  "scopes":[
                     "read",
                     "write"
                  ],
                  "topic":"pend13-data"
               },
               "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b3BpYyI6InBlbmQxMy1kYXRhIiwicHJlZml4Ijoic2Vzc2lvbiIsInNjb3BlcyI6WyJyZWFkIiwid3JpdGUiXSwiYXVkIjpbImh0dHBzOi8vcmVsYXktYWNjZXNzLnByYWN0YWJsZS5pbyJdLCJleHAiOjE2NzA3MDMzNDQsIm5iZiI6MTY3MDcwMzA0NCwiaWF0IjoxNjcwNzAzMDQ0fQ.6bbZoNK-dCrRP_I-m6D1KQmOZ0_uHOaJlBnaAcI5trI",
               "url":"https://relay-access.practable.io/session/pend13-data",
               "verb":"POST"
            },
            {
               "for":"video",
               "permission":{
                  "audience":"https://relay-access.practable.io",
                  "connection_type":"session",
                  "scopes":[
                     "read"
                  ],
                  "topic":"pend13-video"
               },
               "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b3BpYyI6InBlbmQxMy12aWRlbyIsInByZWZpeCI6InNlc3Npb24iLCJzY29wZXMiOlsicmVhZCJdLCJhdWQiOlsiaHR0cHM6Ly9yZWxheS1hY2Nlc3MucHJhY3RhYmxlLmlvIl0sImV4cCI6MTY3MDcwMzM0NCwibmJmIjoxNjcwNzAzMDQ0LCJpYXQiOjE2NzA3MDMwNDR9.NUvssbuUHIJNJ82AzExff2Rx3ckN2jH1lTQLAu_DsEs",
               "url":"https://relay-access.practable.io/session/pend13-video",
               "verb":"POST"
            }
         ],
         "uis":[
            {
               "description":{
                  "further":"https://static.practable.io/info/penduino-basic-ui-1.0/index.html",
                  "image":"https://assets.practable.io/images/booking/ui/penduino-default-1.0/image.png",
                  "long":"Start the pendulum swinging, and observe the difference in decay time for four different types of swinging - driving, braking, short circuit load, and open circuit coil.",
                  "name":"Penduino (Default)",
                  "short":"Control an Electromagnetic Pendulum",
                  "thumb":"https://assets.practable.io/images/booking/ui/penduino-default-1.0/thumb.png",
                  "type":"penduino-default-ui-1.0"
               },
               "streamsRequired":[
                  "data",
                  "video"
               ],
               "url":"https://static.practable.io/ui/penduino-1.0/?streams={{streams}}&exp={{exp}}"
            }
         ]
      }
   ],
   "max":2,
   "msg":"For support contact timothy.drysdale@ed.ac.uk"
}
```

The first stream token is
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b3BpYyI6InBlbmQxMy1kYXRhIiwicHJlZml4Ijoic2Vzc2lvbiIsInNjb3BlcyI6WyJyZWFkIiwid3JpdGUiXSwiYXVkIjpbImh0dHBzOi8vcmVsYXktYWNjZXNzLnByYWN0YWJsZS5pbyJdLCJleHAiOjE2NzA3MDMzNDQsIm5iZiI6MTY3MDcwMzA0NCwiaWF0IjoxNjcwNzAzMDQ0fQ.6bbZoNK-dCrRP_I-m6D1KQmOZ0_uHOaJlBnaAcI5trI
```
```
{
  "topic": "pend13-data",
  "prefix": "session",
  "scopes": [
    "read",
    "write"
  ],
  "aud": [
    "https://relay-access.practable.io"
  ],
  "exp": 1670703344,
  "nbf": 1670703044,
  "iat": 1670703044
}
```

## Gotchas

### Re-defining the Manifest data structure in swagger

There were unexpected errors when letting the swagger boilerplate parse the manifest as `application/yaml` in the request body, such as missing some of the sub fields like `topic_stub` that were very definitely in the manifest, e.g. for ReplaceManifest. There are no yaml annotations in the boiler plate models. Attempting to parse the manifest as `application/json` instead, which have the annotations, still threw the same errors. Initial fault-finding was not particularly illuminating as to where the issue lay, so the workaround is to avoid the issue entirely and send the manifest as a `plain/text` string and parse within the handler. This is more efficient and easier to maintain because it avoids an unneccesary conversion to `models.Manifest` and then to `store.Manifest` and back again on export.

### Time
[time](https://pkg.go.dev/time#Time)
"Representations of a Time value saved by the GobEncode, MarshalBinary, MarshalJSON, and MarshalText methods store the Time.Location's offset, but not the location name. They therefore lose information about Daylight Saving Time."




## Out of date material

The material below here is from early stages of the development, and may not be consistent with the current implementation. 

TODO - edit the below to match current implementation and shift above into main section  where appropriate, or delete otherwise.


### A note on filters and redundancy

Why not have filters on `slots` AND `policies`?  In order to provide access to an experiment, there must be available resources. Forcing all knowledge about availability of resources into a single place (slot definitions) simplifies testing/reporting, at the cost of having more `slots`. 

Class E requires Exclusive access to a Experiement A. Class H will be heavy users of Experiment A. The General Public will also use Experiment A, but only when it is not in high-demand for Class H (e.g. before/after class H use it, and during anti-social hours).

Three slots are prepared for each kit instance of Experiment A, representing (1) the exclusive access periods for Class E, (2) the allowable access periods for Class H, (3) and the allowable access periods for General Public.

To see whether a resource is in exclusive or shared use at any given time, plot the allowable access periods for each resource, and annotate with the slot names.




AVL trees are used to ensure good average look-up performance.

## Features

// it would be helpful if people were warned about non-redundant equipment
// e.g. messaging them when equipment is known to be offline and unavailable for their booking
// or alerting them to a status page.

// Granularity of booking, and display of booking slots.
// we can let the user interface invert bookings information to show availability.
// If we assume "bookable unless booked", then graphically, put the background colour
// to the "free" colour, and overlay bookings in "booked" colour. Save processing it.


## Gitops

In the last booking system, updates could be made on the manifest by mutating it, but this was never used in practice, because the manifest was deleted and reloaded every time we made a change. So we can reduce a lot of unused code if we do away with the live mutation of the manifest, and take a more gitops approach. 

We could host the manifest in a repo, and simply point the booking system at that repo, and have it notice when it has changed (and reload).

Bookings, if saved to disk, could be reloaded. A time-stamped transaction log would do - then any thing that could be re-booked could be flagged during the replay process, e.g. if the slot count is reduced. A user could be alerted by putting that booking into the cancelled bookings list ...  with a flag that the admin cancelled it. 



### Frontend to the existing immediate use booking system

This section is deprecated.

Rather than try to manage all the equipment, this system is simply about assigning users to slots, which are promises of access to kit. Since the kit cannot be accesed through any other gatekeeper, this booking system simply waits for pre-booked slots to start, then obtains a token from the underlying instant use booking system. The system is reliant on the exclusive connection to the underlying system:

```mermaid
graph TD
    U[User] --> A;;	
	A[Booking App] -->  C[New advance booking system] ;;
    C --> E[Existing immediate-use booking system];;
	
```

In the case where multiple booking systems are sharing that access, then there needs to be an interval booking system in front of the instant access system, to mediate between the two competing booking services.


```mermaid
graph TD
    U[User] --> A;;	
	A[Booking App] --> C[New advance booking system] ;;
	
    U2[User#2] --> A2;;	
	A2[Booking App#2] --> C2[New advance booking system#2] ;;

	C --> D[New advance booking system] ;;
	C2 -->	D ;;
    D --> E[Existing immediate-use booking system];;	
```	


### Export / Import

Booking in advance requires persisting state. This can be done in two main ways

0. export/import the internal representation
0. export/import the transactions used to develop in the internal state

The internal representation might change, e.g. to support new features. Or, there may be an issue with the state becoming corrupted, or there may be need to be a change in the number of booking slots available. Therefore a transaction based approach is preferred. There is perhaps additional overhead after import, however it provides greater flexibility for the future.

We will use this feature for testing. A couple of additional points:
- we'll need to know whether a booking was taken up, so keeping track of booking take ups may be a useful thing to do in this system.
- in fact, it is essential, because if a user cancels, another user books in their place, then we restart, we need to restore the cancellation so the replacement booking is valid when booking system is reconstituted.
- we might have different needs around the storage of the transaction history, so we need to pass the storage unit to the system. It needs to implement an "add record" function.
- we could do the import with a similar system, that provides a get record function. Perhaps a "withimport" function would help at startup. 


### Cancellation

With JWT token authorisation, and no revocation, meaningful cancellation is not possible (a user might cancel then use their browser history to return to the session, interfering with the next user taking over from them).  However, sessions booked in advance, and not yet taken up, have not had their JWT token released, so can be safely cancelled.

in the old `pools/activity.go`:
```go
	return permission.Token{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience: jwt.ClaimStrings{*mp.Audience},
		},
		ConnectionType: *mp.ConnectionType,
		Scopes:         mp.Scopes,
		Topic:          *mp.Topic,
	}
```

We could shift to using a `jti` field for each token, so that every time you want to reconnect, you need to get a token from the booking system .... 

edge case - did you get rejected because of no connection, or because the token was invalid?
What load does it put on the booking system?

But - it would help us manage cancellation, potentially. Because we just kill all connections with a certain booking id, then don't make any more tokens for it. Then no existing token can be re-used, so long as the jti list remains intact.

We could also deny-list cancelled bid.

There would be a smaller list to search if we just closed existing connections and deny-listed cancelled bid. We also would not need to revise the current UI because the bid would be in the token.

Better as an explicit claim of `bid` rather than mis-use `jti`



### Identity

Without knowing a user's identity from an authoratitive source, we rely on reasonably co-operative users to respect the operation of the pseudoanonymous UUID we place in their browser cache. If they book on one device and want to use it on another, they can always copy across the link in the actual session (e.g. via messaging, email - the links are long and would not be convenient to hand transcribe). Therefore, there is no need to provide a booking transfer system, because it does not add any particular value, is open to spamming (e.g. randomly trying booking codes).


### Configuration

Systems with advance booking can choose to offer fewer slots than they have equipment for, so as to provide some redundancy. For some experiments, there are no duplicates. There could be quite complicated ways in which booking decisions might need to be made, beyond redundancy e.g. assign any equipment to a person on their first visit, and the same equipment again on any subsequent visit. Therefore, concerns around how to book could usefully be separated from the existing booking service. This way, new policies can be developed and implemented without modifying existing booking systems. This implies a front-end in front of the different booking systems:

```mermaid
graph TD
    U[User] --> A;;	
	A[Booking App] --> B[Booking Proxy] ;;
	B --> C[Allocated/Advance] ;;
	B --> D[Freechoice/Advance] ;;
	B --> E[Freechoice/Instant] ;;
    C --> F[Main #1] ;;
    D --> G[Main #2] ;;
    E --> F ;;
	
	```

The booking app has to know about the different forms of requests it needs to be able to make to each different type of booking system it is interacting with. For each pool that can be booked, the app must know the booking policy that applies. There needs to be a defined way to offer different booking types to users. For example:
Policy could be something like 

### immediate/fixed
current system, immediate bookings only, offer four different length sessions

### immediate/flexible
immediate bookings only, but the length of the booking can be defined with arbitrary granularity from a given minimum booking length, to maximum length. We need to limit minimum booking length to avoid surprises like booking a one-second slot, waiting for any checks, and then being thrown straight back to the UI. different minimum lengths would apply to different users, so that system agents such as checkers could make shorter bookings than would be meaningful for a user. So policies for users must fall within the constraints of the equipment's policy limits. E.g. policy limit on equipment might be three hour bookings, but a certain group of users are given a policy of up to 90minutes bookings only. 


### advance/fixed

Bookings can be made either now, or starting at a later time, within a maximum window, with a fixed set of session lengths on offer. acceptable start times would fall on a repeating pattern, to avoid wastage

## advance/flexible
Bookings can be made either now, or starting at a later time, within a maximum window, with an arbitrary length session (less than a maximum allowed), starting at an arbitrary time. Efficiency of bookings is left to users to arrange, e.g. by sliding the start/end of their bookings up against other bookings.

### Some other policy ....

If there are elements that a user must select, these could be added to the booking form via an automatic field generation approach - describe what needs to be asked, and have the booking system ask it. This would allow new booking policies to be developed without necessarily requiring the booking app to be updated.

The date/time picker would be added with some parameters to control granularity, start times, max time. Also, to show any bookable/non-bookable times.



## Polciy updates on the fly

It would be helpful to be able to extend/retract the booking window without having to do a server restart/booking transaction reload. For example, initial rollout might allow a 24hour booking window. If this seems to be working, the window can be extended to 48hours, then a week etc. If it is retracted, any bookings already made will probably be kept.

Would be helpful to be able to clean out bookings as well - e.g. if we get spammed. This is a prompt to remember rate-limiting.





# Interval -- README from internal interval package

## Motivation

This implementation of the interval tree data structure is intended to work with arbitrary intervals of time, represented by integers. This retains the generality of the unix epoch approach to representing moments in time, with the resolution of the epoch time being left open to either seconds, milliseconds etc as required by the user of the library. These intervals of time are represented on a one-dimensional interval tree. 

### Limitation to one-dimension
The current implementation is limited to one-dimension because this represents the minimum viable approach required. While there is a second potential dimension to be considered, which is multiple instances of the same fungible item being booked - that is out of scope of the present implementation because iterating over all items of equipment in a list to check whether they are free or not, requires at most only one available item to be returned, and not a simultaneous search for all available kits. The use case for finding all available kits is in visualising the future availability of all kits, but this is probably efficiently solved by querying availability at individual points in time as appropriate to the graph (e.g. at half-hourly points for the next 12 hours). 

### Motivation

Booking systems typically simplify their data structures by enforcing opinions about the allowable booking slots. For example, hotel rooms are booked by the night, and gym classes are booked by the hour. Remote laboratory experiments differ in that assumptions made for one type of experiment, can be vastly different for another. Making a one-size-fits-no-one assumption about what slot size to pick might be an acceptable tradeoff in a small laboratory with relatively homogenous experiments, but soon leads to tensions when the laboratory expands, or encounters suggestions for new use cases not previously anticipated. Let's consider a limited selection of some already-known experiment types to see where the issues lie.

### Example durations for some typical experiments

Experiments can have vastly different run times depending on the nature of the experiment, and the educational task. For example:

- a wobbling beam experiment might require only 15 seconds to run, in a batch job. 
- a pendulum exploration exercise might take 15 minutes
- a truss experiment might need 20 minutes
- a spinning disk experiment might need 90 minutes 

These same experiments may also need maintenance windows, or self-check windows, where they cannot be booked by users. A self check routine can take anywhere from around a few seconds for confirming the video and audio are working, to several minutes if physical elements need to stabilise or there are multiple measurements that need to be checked.

### Previous booking system implementations

Booking systems for remote laboratories have tended to mirror classroom schedules, with bookable hour-long slots starting at the top of the hour. In the case of a laboratory with a self-check task, the booking can be shrunk to 55 minutes, with a 5 minute window for the self checking.

### Issues with fixed schedules

Most of the issues relate to scaling up to multiple locations and multiple types of equipment, and experimental tasking. For example:

- task durations vary depending on the educational use case E.g. open-days might need only short slots to get maximum throughput for an introductory exercise in a limited time, while a later year undergraduate may require long slots for in-depth working. 

- lecture demonstrations often come at the start or the end of the lecture, and conference presentations can happen at any time. So you can't shut the lab down for a self check at a fixed time every hour without causing inconvenience. I've previously had to schedule demonstrations around such self-check windows, but do not think organisers of events, or indeed regular users, would tolerate this once it is a mainstream activity.

- multi-location campuses often ofset lecture times to accommodate students travelling from one campus to another, e.g. University of Edinburugh lecture times shift from starting at the top of the hour, to sta# interval
This package implements bookings with arbitrary durations

## Features

- Advance bookings 
- Arbitrary intervals for bookings
- Guarantee exclusive bookings
- Different UIs for different users
- Different limits on number of current/future bookings for different users
- Different usage limits for different users
- Resource status tracking (e.g. if offline for maintenance or failing tests)

## Limitations
- Redundancy is not supported - it's doable, but too complicated to reason about hiearahical groupings of kit at this stage where some people book a type of experiment and others book specific examples.
- Unlimited access to simulations (unlimited, anytime, no need for booking?) - relative simple, but out of scope for now (simple mod for later).
- What if kit doesn't work - how to get another? If cannot cancel ...?! For now - when taking up booking, it returns, equipment not available. More advanced options can be added later (e.g. admin edits bookings to substitute another piece of kit instead, when taking an experiment offline).

## Implementation overview

### Definitions

A `resource` represents a single bookable entity which is a real, physical piece of equipment. The entity can be booked for any arbitrary interval, so long as it does not overlap with any other booking. Booking exclusivity is guaranteed at the resource level. A resource has an availability (bool) and a status message (e.g. "Test failed on (date)", or "Test passed on (date)")

A `filter` represents a collection of `allowed` and `denied` periods, allowing for the implementation of policies about when a booking can be made.

A `slot` represents access to one resource, where the access is restricted by a `filter`. There are potentially many slots for one resource, and slots may overlap, resulted in shared access (but the first to book a resource gets exclusive use of it, it's just the opportunity to book it that is shared). The slot lists the user interface set that can be used with the hardware.

A `user` represents an entity that can book `slots` according to zero or more `policy` instances.

A `policy` represents the maximum usage permitted for a list of one or more slots, the minimum or maximum length of bookable interval, the maximum number of current/future bookings.

#### What happened to pools?

Now that we are booking in advance, the concept of redundancy as it applies to different hierarchies of groupings of equipment is complex, and so pools and redundancy is left for future work. Instead for now, if a user books something, and it becomes unavailable, then they can rebook on something else. It may be upsetting to a user to have a session cancelled due to kit failure, and having to sort a replacement session themselves, but better we reduce frustration overall for many users by offering some form of booking.

#### Why is the filter in the slot, not the policy?

Some resources may be shared between two courses, other between three or four courses, and some exclusive users may only need a subset of the equipment of that type, so blanket filtering of access to a set of resources is not sufficient. Slots with individually-defined access times to specific resources can be added to a policy to express this, for example half the slots for the other class will be available any time, the other half only when not being used by the exclusive user. In other words, the slots in a policy may have different filters in them.  This does require editing all slots related to a resource to establish a new exclusive user, but that can be avoided later by using a rule-based system inspired by or based on Ory Keto / Google Zanzibar.

### Example Process Flow 

A course organiser and practable administrator agree a `policy` for use by the students in a class, including the times of access to equipment which is set by the slots assigned to the policy. Optionally additional policies for course staff and tutors/demonstrators could be agreed, e.g. with access to other user interfaces.

The policy sets user-specific aspects such as total usage, booking duration and live booking count. The slots contain any limitatations on when the equipment is available to a class. Thus, the policy, once agreed need not be changed, however, the definition of the slots it refers to, may change e.g. if another class needs the equipment exclusively for an hour, then that hour would be denied in the slots assigned to other classes. Such changes would happen behind the scenes, and not be explicitly shown in the policy.

A policy exists as a JWT token on a server. A user is provided with a link to a booking system with the policy-code as a URL param, as we currently do for setting groups.

A booking agent (client-side javascript, usually) checks to see if it is has already used the service before, by looking in its cache for pseudo-id. If no pseudo-id is found, it asks the user for their exam code, and checks/validates the format of the exam-code.

The user will see any bookings they have made under any other other policy, but can only book under the policy they currently have - they have to go to the link for the other course to book for that course. In this way, we can refresh the policy as needed, rather than having old policy codes hanging around or storing them in the booking system?  Simpler to test if we don't remember codes? Or store policy codes with expiration date? Get the policy fresh everytime anyway .... What if a code is shared widely? We can just change the link. So easier if we don't save old policies, because then we do not have to tell the booking system to invalidate old policies stored with users - unless we can auto-remove them if they are found to be expired?  These complexities can be better handled when we have the Ory version up and running, so don't get bogged down in it for now - just go with the live-only/non-stored policies for simplicity.

The policy is submitted along with any requests for availability or bookings, as the authorisation.

If a user has spare bookings available then can make a new booking. They can see availability, before cancelling their current booking(s), to make new ones. Don't offer a booking swap at this stage?

Cancel bookings is an option.

Store booking against the pseudo-id so it can be retrieved on future visits.

When a booking becomes due, submit the booking to the server for streaming stuff, same as before.

In this first version, don't allow session cancellation / i.e. can't cancel booking after launching it, to save having to re-do the user interfaces at this stage.


#### Implementation of booking modes

The following parameters have been added to store.Policy

support simulations
```
EnforceUnlimitedUsers bool 
```

support instant use only
```
EnforceStartsWithin bool
StartsWithin time.Duration 
```

Give some latitude if there are delays in processing requests for bookings starting "now"
```
EnforceAllowStartInPast bool
AllowStartInPastWithin time.Duration
```

support next available mode
```
EnforceNextAvailable    bool 
NextAvailable time.Duration 
```

in api spec format:
```
  allow_start_in_past_within:
    type: string
  enforce_allow_start_in_past:
    type: boolean
  enforce_next_available:
    type: boolean
  enforce_starts_within:
    type: boolean
  enforce_unlimited_users:
    type: boolean
  next_available:
    type: string
  starts_within:
    type: string

```


There are no pools now - this needs updating 

```mermaid
erDiagram
    POOL ||--o{ EQUIPMENT: accesses
    EQUIPMENT {
        string name
    }
    EQUIPMENT ||--|{ AVAILABILITY_POLICY : using 
    AVAILABILITY_POLICY {
        string name
    }
    AVAILABILITY_POLICY ||--o{ ALLOWED_INTERVAL : contains
    ALLOWED_INTERVAL {
        time nbf
        time exp
    }
    AVAILABILITY_POLICY ||--o{ DENIED_INTERVAL : contains
    DENIED_INTERVAL {
        time nbf
        time exp
    }


```
tateDiagram-v2
    [*] --> LocalStoreUserName 
    LocalStoreUserName --> FetchUserName
    FetchUserName --> AwaitUserNameResponse
    AwaitUserNameResponse --> UserNameResponseError
    AwaitUserNameResponse --> UserNameResponseOK
    UserNameResponseError --> UserNameBackoffWait
    UserNameBackoffWait --> FetchUserName
    UserNameResponseOK --> FetchLoginToken
    FetchLoginToken --> AwaitLoginTokenResponse
    AwaitLoginTokenResponse --> LoginTokenResponseError
    AwaitLoginTokenResponse --> LoginTokenResponseOK
    LoginTokenResponseError --> LoginTokenBackoff
    LoginTokenBackoff --> FetchLoginToken
    LocalStoreUserName -->  FetchLoginToken
    FetchLoginToken --> FetchGroups
    FetchGroups --> AwaitFetchGroupsResponse
    AwaitFetchGroupsResponse --> FetchGroupsResponseError
    AwaitFetchGroupsResponse --> FetchGroupsResponseOK
    FetchGroupsResponseError
    GetGroups --> Wait
	Wait --> GetGroupPolicies
    GetGroupPolicies --> GetAvailability
    GetAvailability --> CalculateSlots
    CalculateSlots --> Wait
    WaitRequestBooking
    RequestBooking --> AwaitResponse
    AwaitResponse --> OK
    AwaitResponse --> Error

    
    Crash --> [*]
```
