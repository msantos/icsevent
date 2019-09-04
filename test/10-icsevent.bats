#!/usr/bin/env bats

@test "icsevent: all events" {
  run icsevent --start=1564646399 --duration="$((30*24))h" file:///$PWD/test/basic.ics

  expect='1564646400 1 start event+1 End+event+time+overlaps+with+event+with+timezone. -
1564650000 3601 end event+1 End+event+time+overlaps+with+event+with+timezone. -
1564650000 3601 start event+with+timezone Start+time+overlaps+with+event+1. Toronto
1564657200 10801 end event+with+timezone Start+time+overlaps+with+event+1. Toronto
1565652600 1006201 start event+2 Begin+description.%5Cn%5CnBody+of+the+description%2C+wrapping+at+74+characters.+End+description. location+of+event
1565656200 1009801 end event+2 Begin+description.%5Cn%5CnBody+of+the+description%2C+wrapping+at+74+characters.+End+description. location+of+event
1565769600 1123201 start event+3 - -
1565776800 1130401 end event+3 - -
1565823600 1177201 start recurring+event - -
1565841600 1195201 end recurring+event - -'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
}
 
@test "icsevent: all events: exact start time" {
  run icsevent --start=1564646400 --duration="$((30*24))h" file:///$PWD/test/basic.ics

  expect='1564646400 0 start event+1 End+event+time+overlaps+with+event+with+timezone. -
1564650000 3600 end event+1 End+event+time+overlaps+with+event+with+timezone. -
1564650000 3600 start event+with+timezone Start+time+overlaps+with+event+1. Toronto
1564657200 10800 end event+with+timezone Start+time+overlaps+with+event+1. Toronto
1565652600 1006200 start event+2 Begin+description.%5Cn%5CnBody+of+the+description%2C+wrapping+at+74+characters.+End+description. location+of+event
1565656200 1009800 end event+2 Begin+description.%5Cn%5CnBody+of+the+description%2C+wrapping+at+74+characters.+End+description. location+of+event
1565769600 1123200 start event+3 - -
1565776800 1130400 end event+3 - -
1565823600 1177200 start recurring+event - -
1565841600 1195200 end recurring+event - -'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
}
 
@test "icsevent: all events: second after start" {
  run icsevent --start=1564646401 --duration="$((30*24))h" file:///$PWD/test/basic.ics

  expect='1564650000 3599 end event+1 End+event+time+overlaps+with+event+with+timezone. -
1564650000 3599 start event+with+timezone Start+time+overlaps+with+event+1. Toronto
1564657200 10799 end event+with+timezone Start+time+overlaps+with+event+1. Toronto
1565652600 1006199 start event+2 Begin+description.%5Cn%5CnBody+of+the+description%2C+wrapping+at+74+characters.+End+description. location+of+event
1565656200 1009799 end event+2 Begin+description.%5Cn%5CnBody+of+the+description%2C+wrapping+at+74+characters.+End+description. location+of+event
1565769600 1123199 start event+3 - -
1565776800 1130399 end event+3 - -
1565823600 1177199 start recurring+event - -
1565841600 1195199 end recurring+event - -'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
}

@test "icsevent: recurring events" {
  run icsevent --start=1565776801 --duration="$((12*30*24))h" file:///$PWD/test/basic.ics

  expect='1565823600 46799 start recurring+event - -
1565841600 64799 end recurring+event - -
1573689600 7912799 start recurring+event - -
1573707600 7930799 end recurring+event - -
1581552000 15775199 start recurring+event - -
1581570000 15793199 end recurring+event - -
1589410800 23633999 start recurring+event - -
1589428800 23651999 end recurring+event - -'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
}

@test "icsevent: wait: recurring event" {
  run icsevent --dryrun --wait --start=1565776801 --duration="$((12*30*24))h" file:///$PWD/test/basic.ics

  # date --date=@$((1565776801+46799))
  # Wed Aug 14 19:00:00 EDT 2019
  expect='sleep: 46799
Wed Aug 14 19:00:00 EDT 2019: start: recurring event'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
}

@test "icsevent: wait: next event" {
  run icsevent --dryrun --wait --start=1564646399 --duration="$((30*24))h" file:///$PWD/test/basic.ics

  expect='sleep: 1
Thu Aug  1 04:00:00 EDT 2019: start: event 1
Description: End event time overlaps with event with timezone.'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
}

@test "icsevent: wait: wait for next event" {
  run icsevent --wait --start=1564646399 --duration="$((30*24))h" file:///$PWD/test/basic.ics

  expect='Thu Aug  1 04:00:00 EDT 2019: start: event 1
Description: End event time overlaps with event with timezone.'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
}

@test "icsevent: wait: exact start time" {
  run icsevent --dryrun --wait --start=1564646400 --duration="$((30*24))h" file:///$PWD/test/basic.ics

  expect='sleep: 0
Thu Aug  1 04:00:00 EDT 2019: start: event 1
Description: End event time overlaps with event with timezone.'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
}

@test "icsevent: wait: sleep for exact start time" {
  run icsevent --wait --start=1564646400 --duration="$((30*24))h" file:///$PWD/test/basic.ics

  expect='Thu Aug  1 04:00:00 EDT 2019: start: event 1
Description: End event time overlaps with event with timezone.'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
}

@test "icsevent: wait: exact end time" {
  run ./icsevent --wait --start=1565656200 --duration="$((30*24))h" file:///$PWD/test/basic.ics

  expect='Mon Aug 12 20:30:00 EDT 2019: end: event 2
Location: location of event
Description: Begin description.\n\nBody of the description, wrapping at 74 characters. End description.'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
}

@test "icsevent: wait: poll interval: minimum wait" {
  run icsevent --dryrun --wait=false --wait-min=900 --start=1464646401 --duration="15m" file:///$PWD/test/basic.ics

  expect=''

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
  [ "${status}" -eq 0 ]
}

@test "icsevent: wait: poll interval: maximum wait" {
  run icsevent --dryrun --wait=true --wait-max=60 --start=1564650001 --duration="15m" file:///$PWD/test/basic.ics

  expect='sleep: 60'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
  [ "${status}" -eq 0 ]
}

@test "icsevent: output format" {
  run icsevent --output-format="{{ printf \"%s: %s\n\" .Status .Summary }}" \
               --start=1564646399 \
               --duration="$((30*24))h" \
               file:///$PWD/test/basic.ics

  expect='start: event 1
end: event 1
start: event with timezone
end: event with timezone
start: event 2
end: event 2
start: event 3
end: event 3
start: recurring event
end: recurring event'

cat << EOF
expect
======
$expect

output
======
$output
EOF

  [ "${output}" = "$expect" ]
  [ "${status}" -eq 0 ]
}
