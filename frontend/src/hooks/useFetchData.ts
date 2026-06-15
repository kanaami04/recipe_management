import type { Token } from '@/type/LoginUser'
import type { RecipeDataType, RecipeLabel, SharedUser } from '@/type/RecipeDataType'

import { useApiData } from './useApiData'

export const useFetchSharedUser = (token: Token) => {
  return useApiData<Array<SharedUser>>('api/users/', token)
}

export const useFetchRecipeLabel = (token: Token) => {
  return useApiData<Array<RecipeLabel>>('api/label/', token)
}

export const useFetchRecipes = (token: Token) => {
  return useApiData<Array<RecipeDataType>>('/api/recipes/', token)
}

export const useFetchData = <T>(url: string, token: Token) => {
  return useApiData<T>(url, token)
}
