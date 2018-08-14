Feature: As a developer, I want to ensure that the 
  loabalancer functions correctly with basic Consul prepared queries

  Scenario: Calls one upstream
    Given that Consul is running
      And I have a prepared query
      And 1 services are started
    When I call the client 10 times with a query
    Then I expect 1 different endpoints to have been called

  Scenario: Calls two different upstreams
    Given that Consul is running
      And I have a prepared query
      And 2 services are started
    When I call the client 10 times with a query
    Then I expect 2 different endpoints to have been called
