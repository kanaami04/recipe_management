import type { RecipeDataType, SharedUser, RecipeLabel } from "@/type/RecipeDataType"
import type { LoginUser } from "@/type/LoginUser"
import { useApiData } from "./useApiData"
import type { Token } from "@/type/LoginUser"

export const useFetchSharedUser = (token: Token) => {
  return useApiData<Array<SharedUser>>("api/users/" ,token)
}

export const useFetchRecipeLabel = (token: Token) => {
  return useApiData<Array<RecipeLabel>>("api/label/", token)
}

export const useFetchLoginUser = (token: Token) => {
  return useApiData<LoginUser>("api/user_info/", token)
}

export const useFetchRecipes = (token: Token) => {
  return useApiData<Array<RecipeDataType>>("/api/recipes/", token)
}

export const useFetchData = <T>(url: string, token: Token) => {
  return useApiData<T>(url, token)
}