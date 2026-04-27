import { isAuthenticated } from "@/auth/guards";
import { Center, Loader, Stack, Text } from "@mantine/core";
import { createFileRoute, redirect } from "@tanstack/react-router";
import { valibotValidator } from "@tanstack/valibot-adapter";
import { signInWithRedirect } from "aws-amplify/auth";
import * as v from "valibot";

const SignInSearch = v.object({
  redirect_to: v.optional(v.string()),
});

export const Route = createFileRoute("/_public/sign-in")({
  validateSearch: valibotValidator(SignInSearch),
  beforeLoad: async ({ search }) => {
    if (await isAuthenticated()) {
      throw redirect({ to: search.redirect_to ?? "/" });
    }
    await signInWithRedirect({
      customState: search.redirect_to,
      options: { lang: "ja" },
    });
  },
  component: SignIn,
});

function SignIn() {
  return (
    <Center h="100vh">
      <Stack align="center" gap="md">
        <Loader />
        <Text c="dimmed">ログイン画面へリダイレクト中</Text>
      </Stack>
    </Center>
  );
}
