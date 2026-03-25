/**
 * Icon mapping constants for TelePromptr.
 *
 * Centralizes icon assignments for navigation items, entity types,
 * status indicators, and common actions. Import individual icons
 * to ensure tree-shaking removes unused ones from the bundle.
 *
 * @module
 */

// Navigation icons
export {
	Activity as TracesIcon,
	ClipboardCheck as EvaluationsIcon,
	FileText as PromptsIcon,
	BarChart3 as AnalyticsIcon,
	Settings as SettingsIcon
} from 'lucide-svelte';

// Status icons
export {
	CheckCircle as SuccessIcon,
	AlertTriangle as WarningIcon,
	XCircle as ErrorIcon,
	Info as InfoIcon
} from 'lucide-svelte';

// Action icons
export {
	Plus as PlusIcon,
	Pencil as EditIcon,
	Trash2 as DeleteIcon,
	Copy as CopyIcon,
	Download as DownloadIcon,
	Filter as FilterIcon,
	ArrowUpDown as SortIcon,
	Search as SearchIcon,
	RefreshCw as RefreshIcon,
	MoreHorizontal as MoreIcon,
	ExternalLink as ExternalLinkIcon
} from 'lucide-svelte';

// Layout / UI icons
export {
	PanelLeftClose as SidebarCollapseIcon,
	PanelLeftOpen as SidebarExpandIcon,
	Menu as MenuIcon,
	X as CloseIcon,
	ChevronRight as ChevronRightIcon,
	ChevronDown as ChevronDownIcon,
	Sun as LightModeIcon,
	Moon as DarkModeIcon,
	Monitor as SystemModeIcon,
	Bell as NotificationsIcon,
	User as UserIcon,
	LogOut as LogOutIcon,
	ChevronsUpDown as SelectorIcon
} from 'lucide-svelte';

// Entity type icons
export {
	Cpu as ModelIcon,
	Braces as TokenIcon,
	Clock as LatencyIcon,
	DollarSign as CostIcon,
	GitBranch as SpanIcon,
	Tag as TagIcon,
	Layers as VersionIcon
} from 'lucide-svelte';
