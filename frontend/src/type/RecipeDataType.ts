export type RecipeDataType = {
	id?: number;
	created_at?: string;
	updated_at?: string;
	cooking: {
		ingredients: { name: string; };
		quantity: number;
		unit: string;
	}[];
	season: {
		seasoning: { name: string; };
		quantity: number;
		unit: string;
	}[];
	procedure: string;
	owner?: {
		username: string
	};
	shared_user: {
		username: string
	}[] | [];
	title: string;
	create_time: number;
	create_for: number;
	archive_flg: boolean;
	label: {
		name: string;
	}[] | [];
}

export type Material = {
  name: string;
  quantity: number;
  unit: string;
}

export type SharedUser = {
	id?: number;
	username: string;
}

export type RecipeLabel = {
	id?: number;
	name: string;
}