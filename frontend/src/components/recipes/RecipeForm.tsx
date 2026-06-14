import { Button } from "@/components/ui/button"
import {
	DialogFooter,
	DialogClose,
} from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Checkbox } from "@/components/ui/checkbox"
import { RecipeInputForm } from "./RecipeInputForm"
import { Textarea } from "@/components/ui/textarea"
import { useState } from "react"
import type { Material, RecipeDataType } from "@/type/RecipeDataType"
import { SelectInput } from "./SelectInput"
import { MultiSelectInput } from "./MultiSelectInput"

type Props = {
	mode: "create" | "edit";
	initialData?: RecipeDataType;
	onSubmit: (payload: RecipeDataType) => Promise<void>;
	labelData?: { name: string }[];
	sharedUserData?: { username: string }[];
	onClickCancel?: () => void;
}

export function RecipeForm({
	mode,
	initialData,
	onSubmit,
	labelData,
	sharedUserData,
	onClickCancel,
}: Props) {
	const id = initialData?.id ?? null
	const [title, setTitle] = useState(initialData?.title ?? '')
	const [createFor, setCreateFor] = useState<string>(String(initialData?.create_for ?? ''))
	const [time, setTime] = useState<string>(String(initialData?.create_time ?? ''))
	const [ingredients, setIngredients] = useState<Material[]>(initialData?.cooking ? (
		initialData?.cooking.map((value) => (
			{ name: value.ingredients.name, quantity: value.quantity, unit: value.unit }))
	) : [])
	const [seasoning, setSeasoning] = useState<Material[]>(initialData?.season ? (
		initialData?.season.map((value) => (
			{ name: value.seasoning.name, quantity: value.quantity, unit: value.unit }))
	) : [])
	const [procedure, setProcedure] = useState<string>(initialData?.procedure ?? '')
	const [isArchive, setISArchive] = useState(initialData?.archive_flg ?? false);
	const [label, setLabel] = useState<string[]>(initialData?.label?.map((l) => l.name) ?? []);
	const [sharedUser, setSharedUser] = useState<string[]>(initialData?.shared_user?.map((s) => s.username) ?? [])

	const handleCreatePayload: React.FormEventHandler<HTMLFormElement> = async (e) => {
		e.preventDefault()

		const payload = {
			id: Number(id),
			title,
			create_time: Number(time),
			create_for: Number(createFor),
			procedure,
			archive_flg: isArchive,
			label : label ? label.map(l => ({ name: l })) : [] as { name: string }[],
			shared_user: sharedUser ? sharedUser.map(s => ({username: s})) : [] as {username: string}[],
			cooking: ingredients.map((item) => ({
				ingredients: { name: item.name },
				quantity: item.quantity,
				unit: item.unit,
			})),
			season: seasoning.map((item) => ({
				seasoning: { name: item.name },
				quantity: item.quantity,
				unit: item.unit,
			}))
		}

		console.log("送信するJSON:", JSON.stringify(payload, null, 2));
		await onSubmit(payload)
	}

	return (
		<form onSubmit={handleCreatePayload}>
			<div className="grid gap-4 h-140 overflow-auto">
				<div className="flex gap-3">
					<div className="flex-2 grid gap-3">
					<Label>title</Label>
						<Input placeholder="title" value={title} onChange={(e) => setTitle(e.target.value)} />
					</div>
					<SelectInput
						className="flex-1 grid gap-3"
						label="create_for"
						value={createFor}
						onChange={setCreateFor}
						placeholder="number"
						options={[...Array(10)].map((_, i) => ({
							label: `${i + 1}`,
							value: `${i + 1}`,
						}))}
					/>
					<SelectInput
						className="flex-1 w-20 grid gap-3"
						label="create_time"
						value={time}
						onChange={setTime}
						placeholder="time"
						options={[...Array(30)].map((_, i) => {
							const val = (i + 1) * 5.
							return { label: `${val}`, value: `${val}` }
						})}
					/>
				</div>
				<RecipeInputForm label="ingredients" initialInputData={mode === 'create' ? null :ingredients} onChange={setIngredients} />
				<RecipeInputForm label="seasoning" initialInputData={mode === 'create' ? null : seasoning} onChange={setSeasoning} />
				<div className="grid gap-3">
					<Label>procedure</Label>
					<Textarea placeholder="..." value={procedure} onChange={(e) => setProcedure(e.target.value)} />
				</div>
				<div className="flex gap-3">
					{labelData && (
						<MultiSelectInput
							className="flex-1 w-20 grid gap-2"
							label="label"
							value={label}
							onChange={setLabel}
							options={labelData.map((label) => ({
								label: label.name,
								value: label.name,
							}))}
						/>
					)}
					{sharedUserData && (
						<MultiSelectInput
							className="flex-1 w-20 grid gap-2"
							label="shared"
							value={sharedUser}
							onChange={setSharedUser}
							options={sharedUserData.map((user) => ({
								label: user.username,
								value: user.username,
							}))}
						/>
					)}
				</div>
				<div className="flex items-center gap-3">
					<Checkbox id="archive_flg" checked={isArchive} onCheckedChange={(value: boolean) => setISArchive(value)} />
					<Label htmlFor="archive_flg">archive</Label>
				</div>
			</div>
			<DialogFooter>
				<DialogClose asChild>
					<Button variant="outline" onClick={onClickCancel}>Cancel</Button>
				</DialogClose>
				<Button type="submit" >
					{mode === 'create' ? "Create" : "Update"}
				</Button>
			</DialogFooter>
		</form>
	)
}