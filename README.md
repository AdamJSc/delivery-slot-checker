# delivery-slot-checker

This is a program written in Go that runs a periodic check on the availability of UK supermarket delivery slots
(namely Asda), and notifies a set of pre-configured recipients via SMS if any are found.

This was originally built out of necessity during the COVID-19 pandemic, when many people found it nigh-on impossible
to book a delivery slot for their online shopping order given the unprecedented high demand.

## Requirements

* Go 1.13 OR Docker
* Nexmo ([Vonage](https://dashboard.nexmo.com/)) account

## First, the ethical stuff...

This program makes a request directly to Asda's groceries API (used by their own website front-end) in
order to retrieve the latest delivery slots. It filters by those which are available and then uses Nexmo
to issue an SMS with slot details to the mobile numbers stored in a local data file.

**There is NO automation implemented beyond this.**

This program will not create new account sign-ups, nor login and automatically book an available slot etc.
(like a typical "bot" might, for example).

It is intended to serve members, or carers, of the most vulnerable groups who may not have the time or
ability to repeatedly perform this task themselves via the website, and therefore miss out on booking
a much-needed delivery slot for food and other essentials whenever they become available during their
period of isolation.

**If you belong to, or are supporting, one of these vulnerable groups and would like to make use of this service
without any setup, please email me:**

[adam@covidcommunity.info](mailto:adam@covidcommunity.info)

Otherwise, if you are setting it up for your own use...

### PLEASE USE THIS PROGRAM RESPONSIBLY! ###

Running this program is effectively no different to periodically refreshing the Asda website manually (but
without the burden of needing to physically do so).

It has been written to make this check at approximately 10-minute intervals by default, in order to prevent unnecessary spikes in
traffic to the Asda site.

Also, for each postcode this process will be performed each day until at least one available delivery slot has been
retrieved for the first time. Once an SMS has been sent to all recipients, this task is bypassed for a set duration
of time (default is 2 hours), and will resume again after this point.

The interval setting can be changed in the code prior to execution.

**HOWEVER... if you are inclined to amend this, please be a good citizen and consider the implications this will
have.**

(i.e. don't do anything silly and get your IP blocked... 🙂)

## Installation

* Firstly you'll need to create your environment file, which you can do by copying the example provided:

```bash
cp .env.example .env
```

* Next, make sure you have signed up for an account at [Nexmo](https://dashboard.nexmo.com/)

* You'll need to obtain your _API Key_ and _API Secret_, which you can find in the
[Getting Started Guide](https://dashboard.nexmo.com/getting-started-guide), underneath the heading **Your API credentials**.
Add these credentials to your new `.env` file as `NEXMO_KEY` and `NEXMO_SECRET` respectively.

```
NEXMO_KEY=my_key_from_nexmo
NEXMO_SECRET=my_secret_from_nexmo
```

_Please note:_ There is a monetary cost attached to each SMS that is issued via Nexmo. As at April 2020, this is approximately
**0.03-0.04EUR** per message. Please refer to Nexmo's [pricing guide](https://www.vonage.com/communications-apis/sms/pricing/)
for exact costs.

* Now you will need to create your task payloads data file, which you can do by copying the example provided:

```bash
cp data/tasks/payloads.example.yml data/tasks/payloads.yml
```

* In this file, you can configure the postcodes you'd like to find delivery slots for, as well as the recipients who should
receive an alert when available delivery slots are found for each postcode:

```yaml
-
  identifier: ab12-0ab   # log prefix
  interval: 600          # number of seconds between each execution
  postcode: AB120AB      # postcode to search
  recipients:
    -
      name: Mick         # recipient's name
      mobile: +44XXX     # recipient's mobile number
    -
      name: Keith
      mobile: +44XXX
    -
      name: Ronnie
      mobile: +44XXX
    -
      name: Charlie
      mobile: +44XXX
-
  identifier: ba21-9ba
  interval: 600
  postcode: BA219BA
  recipients:
    -
      name: Johnny
      mobile: +44XXX
    -
      name: Joey
      mobile: +44XXX
    -
      name: Dee Dee
      mobile: +44XXX
    -
      name: Tommy
      mobile: +44XXX

```

If you need more postcodes, you can keep copying this data structure and repeat it underneath (as long the file remains valid YAML).

## Usage

To run using native Golang, from the project root:

```bash
go run main.go
```

...or to run using Docker, from the project root:

```bash
docker run --rm -w /go/src/app -v $PWD:/go/src/app golang:1.13 go run main.go
```

## Contributing

It would be cool to integrate some additional supermarkets and means of alerting (email etc.)

Please feel free to fork and open PRs! 😁
