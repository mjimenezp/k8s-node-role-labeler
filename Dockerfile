FROM gcr.io/distroless/static
ADD node-role-labeler /
ENTRYPOINT ["/node-role-labeler"]
