import {
	AlertDialog,
	AlertDialogHeader,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogTitle,
	AlertDialogAction,
	AlertDialogFooter,
} from "@/components/ui/alert-dialog";

type Alert = {
  title: string;
  description: string
  open: boolean;
  onOpenChange: () => void;
}

export function MessageAlertDialog({title, description, open, onOpenChange}: Alert) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription className="whitespace-pre-line">
            {description}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogAction onClick={onOpenChange}>
            OK
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}