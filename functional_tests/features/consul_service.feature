Feature: As a developer, I want to ensure that the 
  loabalancer functions correctly

  Scenario: Calls one upstream
    Given that Consul is running
      And 1 services are started
    When I call use the client 10 times
    Then I expect 1 different endpoints to have been called

  Scenario: Calls two different upstreams
    Given that Consul is running
      And 2 services are started
    When I call use the client 10 times
    Then I expect 2 different endpoints to have been called
  
  Scenario: Handles updates when new services are added
    Given that Consul is running
      And 2 services are started
    When I call use the client 10 times
      And 1 services are started
    When I call use the client 10 times
    Then I expect 3 different endpoints to have been called
  
  Scenario: Handles updates when services are deleted
    Given that Consul is running
      And 2 services are started
    When I call use the client 10 times
      And 1 services are removed
    When I call use the client 10 times
    Then I expect 2 different endpoints to have been called
