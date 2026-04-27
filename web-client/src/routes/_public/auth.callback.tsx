import { Center, Loader, Stack, Text } from "@mantine/core";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { fetchAuthSession } from "aws-amplify/auth";
import { Hub } from "aws-amplify/utils";
import { useEffect, useRef } from "react";

export const Route = createFileRoute("/_public/auth/callback")({
  component: Callback,
});

function Callback() {
  const router = useRouter();
  const targetRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    const goNext = () => {
      router.history.push(targetRef.current ?? "/orgs");
    };

    const unsubscribe = Hub.listen("auth", ({ payload }) => {
      switch (payload.event) {
        case "customOAuthState":
          targetRef.current = payload.data;
          break;
        case "signedIn":
          goNext();
          break;
        case "signInWithRedirect_failure":
          router.history.push("/sign-in");
          break;
      }
    });

    // Race fallback: if Amplify finished URL parse before the listener
    // attached, the signedIn / customOAuthState events would be missed.
    // Re-check the session and navigate if already authenticated.
    fetchAuthSession()
      .then((s) => {
        if (s.tokens) goNext();
      })
      .catch(() => {
        /* not yet signed in, rely on Hub */
      });

    return unsubscribe;
  }, [router]);

  return (
    <Center h="100vh">
      <Stack align="center" gap="md">
        <Loader />
        <Text c="dimmed">ログイン処理中</Text>
      </Stack>
    </Center>
  );
}
