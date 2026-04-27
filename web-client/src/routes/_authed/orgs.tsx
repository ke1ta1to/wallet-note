import { Container, Title } from "@mantine/core";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/orgs")({
  component: Orgs,
});

function Orgs() {
  return (
    <Container py="xl">
      <Title order={1}>Organizations</Title>
    </Container>
  );
}
