export namespace dto {
	
	export class ProjectApiKeyEntry {
	    id: string;
	    project_name: string;
	    project_number: string;
	    api_key: string;
	    enabled: boolean;
	    status: string;
	    cooldown_count: number;
	
	    static createFrom(source: any = {}) {
	        return new ProjectApiKeyEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_name = source["project_name"];
	        this.project_number = source["project_number"];
	        this.api_key = source["api_key"];
	        this.enabled = source["enabled"];
	        this.status = source["status"];
	        this.cooldown_count = source["cooldown_count"];
	    }
	}
	export class AppSettingsDTO {
	    auto_analyze: boolean;
	    primary_model: string;
	    duplicate_guard: boolean;
	    copy_workers: number;
	    active_key_count: number;
	    total_project_count: number;
	    primary_routes: number;
	    total_routes: number;
	    allocated_workers: number;
	    batch_size: number;
	    estimated_rpm: number;
	    estimated_rpd: number;
	    projects: ProjectApiKeyEntry[];
	
	    static createFrom(source: any = {}) {
	        return new AppSettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.auto_analyze = source["auto_analyze"];
	        this.primary_model = source["primary_model"];
	        this.duplicate_guard = source["duplicate_guard"];
	        this.copy_workers = source["copy_workers"];
	        this.active_key_count = source["active_key_count"];
	        this.total_project_count = source["total_project_count"];
	        this.primary_routes = source["primary_routes"];
	        this.total_routes = source["total_routes"];
	        this.allocated_workers = source["allocated_workers"];
	        this.batch_size = source["batch_size"];
	        this.estimated_rpm = source["estimated_rpm"];
	        this.estimated_rpd = source["estimated_rpd"];
	        this.projects = this.convertValues(source["projects"], ProjectApiKeyEntry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileItem {
	    filename: string;
	    size_bytes: number;
	    mod_time: string;
	    full_path: string;
	    is_image: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.size_bytes = source["size_bytes"];
	        this.mod_time = source["mod_time"];
	        this.full_path = source["full_path"];
	        this.is_image = source["is_image"];
	    }
	}
	export class IntegrityResult {
	    total_db: number;
	    total_disk: number;
	    missing_in_disk: string[];
	    matched_count: number;
	    is_exact_match: boolean;
	    alert_message: string;
	
	    static createFrom(source: any = {}) {
	        return new IntegrityResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_db = source["total_db"];
	        this.total_disk = source["total_disk"];
	        this.missing_in_disk = source["missing_in_disk"];
	        this.matched_count = source["matched_count"];
	        this.is_exact_match = source["is_exact_match"];
	        this.alert_message = source["alert_message"];
	    }
	}
	export class MoveSummary {
	    total_files: number;
	    moved_files: number;
	    skipped_files: number;
	    failed_files: number;
	    duration_ms: number;
	    duration_formatted: string;
	    moved_list: string[];
	    skipped_list: string[];
	    failed_list: string[];
	
	    static createFrom(source: any = {}) {
	        return new MoveSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_files = source["total_files"];
	        this.moved_files = source["moved_files"];
	        this.skipped_files = source["skipped_files"];
	        this.failed_files = source["failed_files"];
	        this.duration_ms = source["duration_ms"];
	        this.duration_formatted = source["duration_formatted"];
	        this.moved_list = source["moved_list"];
	        this.skipped_list = source["skipped_list"];
	        this.failed_list = source["failed_list"];
	    }
	}
	export class TagItem {
	    id: number;
	    name: string;
	    count: number;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new TagItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.count = source["count"];
	        this.source = source["source"];
	    }
	}
	export class PhotoDetail {
	    id: number;
	    file_name: string;
	    destination_path: string;
	    full_path: string;
	    thumb_url: string;
	    description: string;
	    tags: TagItem[];
	    size_bytes: number;
	    sha256: string;
	    status: string;
	    backed_up_at: string;
	
	    static createFrom(source: any = {}) {
	        return new PhotoDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.file_name = source["file_name"];
	        this.destination_path = source["destination_path"];
	        this.full_path = source["full_path"];
	        this.thumb_url = source["thumb_url"];
	        this.description = source["description"];
	        this.tags = this.convertValues(source["tags"], TagItem);
	        this.size_bytes = source["size_bytes"];
	        this.sha256 = source["sha256"];
	        this.status = source["status"];
	        this.backed_up_at = source["backed_up_at"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PhotoListItem {
	    id: number;
	    file_name: string;
	    full_path: string;
	    thumb_url: string;
	    description: string;
	    tags: string[];
	    backed_up_at: string;
	
	    static createFrom(source: any = {}) {
	        return new PhotoListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.file_name = source["file_name"];
	        this.full_path = source["full_path"];
	        this.thumb_url = source["thumb_url"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.backed_up_at = source["backed_up_at"];
	    }
	}
	
	export class SearchParams {
	    query: string;
	    tag_ids: number[];
	    page: number;
	    page_size: number;
	    expanded_terms?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SearchParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.tag_ids = source["tag_ids"];
	        this.page = source["page"];
	        this.page_size = source["page_size"];
	        this.expanded_terms = source["expanded_terms"];
	    }
	}
	export class SearchResult {
	    items: PhotoListItem[];
	    total: number;
	    analyzed_total: number;
	    page: number;
	    page_size: number;
	    total_pages: number;
	    expanded_terms?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], PhotoListItem);
	        this.total = source["total"];
	        this.analyzed_total = source["analyzed_total"];
	        this.page = source["page"];
	        this.page_size = source["page_size"];
	        this.total_pages = source["total_pages"];
	        this.expanded_terms = source["expanded_terms"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TagCount {
	    name: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new TagCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.count = source["count"];
	    }
	}
	
	export class VisionResult {
	    description: string;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new VisionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.description = source["description"];
	        this.tags = source["tags"];
	    }
	}

}

export namespace model {
	
	export class Tag {
	    id: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Tag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class PhotoTag {
	    photo_id: number;
	    tag_id: number;
	    source: string;
	    tag: Tag;
	    photo_tags: Photo;
	
	    static createFrom(source: any = {}) {
	        return new PhotoTag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.photo_id = source["photo_id"];
	        this.tag_id = source["tag_id"];
	        this.source = source["source"];
	        this.tag = this.convertValues(source["tag"], Tag);
	        this.photo_tags = this.convertValues(source["photo_tags"], Photo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Photo {
	    id: number;
	    destination_path: string;
	    file_name: string;
	    rel_path: string;
	    size_bytes: number;
	    sha256: string;
	    // Go type: time
	    backed_up_at: any;
	    status: string;
	    description: string;
	    photo_tags: PhotoTag[];
	
	    static createFrom(source: any = {}) {
	        return new Photo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.destination_path = source["destination_path"];
	        this.file_name = source["file_name"];
	        this.rel_path = source["rel_path"];
	        this.size_bytes = source["size_bytes"];
	        this.sha256 = source["sha256"];
	        this.backed_up_at = this.convertValues(source["backed_up_at"], null);
	        this.status = source["status"];
	        this.description = source["description"];
	        this.photo_tags = this.convertValues(source["photo_tags"], PhotoTag);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

