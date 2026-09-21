# Spec Delta

## ADDED Requirements

### Requirement: Bounded hash slicing

`GetCommitInfo` SHALL return a `ShortHash` of length
`min(7, len(hashStr))`, where `hashStr` is the string form of
the commit's full hash. The function SHALL NOT panic when the
hash string is shorter than 7 characters (e.g., a malformed
input); the returned `ShortHash` SHALL be the full string in
that case.

#### Scenario: Normal-length hash returns 7-character short hash
- **WHEN** `GetCommitInfo` is called with a 40-character
  commit hash
- **THEN** the returned `ShortHash` is exactly 7 characters

#### Scenario: Short hash returns the full string
- **WHEN** `GetCommitInfo` is called with a hash string shorter
  than 7 characters
- **THEN** the returned `ShortHash` equals the full input string
  and the function does not panic
