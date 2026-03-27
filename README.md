# keycloak-broker

Note, this is just a prototype. The purpose of this repo is to test a few key assumptions before architecture discussions can happen:
- Keycloak API is fully synchronous
- Keycloak API maps somewhat neatly-ish onto OSB spec or vice-versa
- An OSB can be 100% fully stateless/ephemeral, all state is in Keycloak only
