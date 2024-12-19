# OpenAI ContentGenerator / ThreadHandler

This is a concrete implementation of the minds.ContentGenerator and
minds.ThreadHandler interfaces with OpenAI.

The client implementation is separate than the interface implementation to
allow for easy swapping of the client implementation for testing.