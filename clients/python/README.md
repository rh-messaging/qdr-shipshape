# Python clients to be used within qdr-shipshape

- Config through Property or Environment Variables
  - Common connectivity configuration to be shared with all clients 

- Easy to address custom scenarios and reproducers

- Impacts only Interconnect Tests, so safer to change, than changing
  the regular QE clients (which might impact other teams as well)
  
- Should be able to determine whether or not client succeeded or failed

- Output should provide enough information that can be used to generate a
  valid ResultData instance (for shipshape)
