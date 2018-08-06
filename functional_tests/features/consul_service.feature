Feature: As a developer, I want to ensure that the 
  loabalancer functions correctly

  Scenario: Calls one upstream
    Given that Consul is running
    And the services are running and registered
    When I call use the client 1 times
    Then I expect 1 different endpoints to have been called

  Scenario: Calls two different upstreams
    Given that Consul is running
    And the services are running and registered
    When I call use the client 2 times
    Then I expect 2 different endpoints to have been called
