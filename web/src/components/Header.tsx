import Logo from "./Logo";

export default function Header() {
    return (
        <div className="bg-neutral-50 px-4 py-3">
            <div className="max-w-7xl mx-auto flex justify-between">
                <div className="flex items-center gap-3">
                    <Logo className="w-7" />
                    <h1 className="font-medium text-xl">SmartNotes</h1>
                </div>
            </div>
            <div className="flex"></div>
        </div>
    )
}
