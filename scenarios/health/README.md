## Setup

* One namespace `100-health`
* One broker `testbroker`
* Trigger x 100
* One service per trigger, 100 in total
* One deployment backing the services
  * 10 replicas
  * All pods always return success instantly
* One seeder that sends one event every second with 100 bytes payload