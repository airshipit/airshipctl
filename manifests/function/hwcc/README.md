# Function: hwcc

Controller for classifying host hardware characteristics to expected values.

The HWCC (Hardware Classification Controller) compares and validates the
workload profile against Baremetal Hosts and classifies right match host
and label the host. Also it displays the count for matched, unmatched
and error hosts.

Comparison and validation is done on baremetalhost list provided by `BMO`
against hardware profile mentioned in
`metal3.io_hardwareclassifications.yaml`.

HWCC will label matched hosts.
 * Default

   `hardwareclassification.metal3.io/<PROFILE-NAME>=matches`
 * User Provided

   `hardwareclassification.metal3.io/<PROFILE-NAME>=<LABEL>`

HWCC also label hosts which are in error state if
`hardwareclassification-error=All` label is given in workload profile.

HWCC status shows multiple items w.r.t applied profile :
 * Name of the profile
 * Profile match status
 * Matched Host count
 * Error Host count

## Example Usage

User can validate and classify the hosts based on hardware requirement.
User will get to know how many hosts matched to user profile and
how many hosts are in error state. HWCC status will also show number of hosts
falling under different error states.
User can select any of matched host and go for provisioning.

