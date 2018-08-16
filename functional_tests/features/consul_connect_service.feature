@connect
Feature: As a developer, I want to ensure that the 
  loabalancer functions correctly with Consul Connect services

  Scenario: Calls one upstream
    Given that Consul is running
      And 1 services are started
    When I call the connect enabled client 10 times
    Then I expect 1 different endpoints to have been called

  Scenario: Calls two different upstreams
    Given that Consul is running
      And 2 services are started
    When I call the connect enabled client 10 times
    Then I expect 2 different endpoints to have been called
