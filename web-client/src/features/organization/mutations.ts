import { apiClient } from "@/api/client";
import { mutationOptions } from "@tanstack/react-query";
import * as v from "valibot";

export const NewOrgSchema = v.object({
  name: v.pipe(
    v.string(),
    v.trim(),
    v.minLength(1, "名前は必須です"),
    v.maxLength(100, "100 文字以内で入力してください"),
  ),
});

export type NewOrgInput = v.InferOutput<typeof NewOrgSchema>;

export const createOrgMutation = mutationOptions({
  mutationKey: ["orgs", "create"],
  mutationFn: async (input: NewOrgInput) => {
    const { data, error } = await apiClient.POST("/orgs", { body: input });
    if (error) throw error;
    return data;
  },
});
