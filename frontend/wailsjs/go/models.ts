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

export namespace service {
	
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

}

