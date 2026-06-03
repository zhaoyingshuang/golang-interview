#!/usr/bin/env python3
"""Generate VitePress markdown files from Go web server API."""
import json, os, sys, urllib.request

# Start server, fetch data, generate md files
BASE = "http://localhost:9090"

def main():
    out_dir = sys.argv[1] if len(sys.argv) > 1 else "docs"

    chapter_paths = {
        "01-basics": "basics",
        "02-concurrency": "concurrency",
        "03-memory": "memory",
        "04-performance": "performance",
    }

    chapter_names = {
        "01-basics": "基础篇",
        "02-concurrency": "并发篇",
        "03-memory": "内存篇",
        "04-performance": "性能篇",
    }

    # Fetch topic list
    topics = json.loads(urllib.request.urlopen(f"{BASE}/api/topics").read())

    for t in topics:
        topic_id = t["id"]
        data = json.loads(urllib.request.urlopen(f"{BASE}/api/topic/{topic_id}").read())

        dir_name = chapter_paths[data["chapter"]]
        full_dir = os.path.join(out_dir, dir_name)
        os.makedirs(full_dir, exist_ok=True)

        lines = [
            "---",
            f"title: {data['title']}",
            "---",
            "",
        ]

        for i, section in enumerate(data["sections"]):
            lines.append(f"## {i+1}. {section['title']}")
            lines.append("")

            # Process content text
            content = section["content"]
            # Split at **Keyword** boundaries for paragraphs
            # Insert double newline before ** after sentence endings
            import re
            content = re.sub(r'([。！？])\s*(\*\*)', r'\1\n\n\2', content)
            content = re.sub(r'([；;])\s*(\*\*)', r'\1\n\n\2', content)

            # Format special tags as VitePress containers
            content = re.sub(r'\*\*使用场景\*\*', '::: tip 使用场景\n', content)
            content = re.sub(r'\*\*生产场景\*\*', '::: tip 使用场景\n', content)
            content = re.sub(r'\*\*生产建议\*\*', '::: tip 生产建议\n', content)
            content = re.sub(r'\*\*生产影响\*\*', '::: info 生产影响\n', content)
            content = re.sub(r'\*\*生产选择\*\*', '::: tip 生产选择\n', content)
            content = re.sub(r'\*\*生产坑.*?\*\*', '::: danger 生产陷阱\n', content)
            content = re.sub(r'\*\*面试追问\*\*', '::: warning 面试追问\n', content)
            content = re.sub(r'\*\*面试必记.*?\*\*', '::: warning 面试必记\n', content)
            content = re.sub(r'\*\*面试高频.*?\*\*', '::: warning 面试高频\n', content)
            content = re.sub(r'\*\*面试经典.*?\*\*', '::: warning 面试题\n', content)
            content = re.sub(r'\*\*最佳实践\*\*', '::: tip 最佳实践\n', content)
            content = re.sub(r'\*\*正确做法\*\*', '::: tip 正确做法\n', content)
            content = re.sub(r'\*\*为什么需要.*?\*\*', lambda m: f'::: info {m.group()[2:-2]}\n', content)

            # Close open containers at paragraph breaks
            content = re.sub(r'\n\n(?!::)', '\n:::\n\n', content)

            # Handle remaining **bold** (normal bold)
            # Already valid markdown

            lines.append(content)
            lines.append("")

            if section.get("code"):
                lines.append("```go")
                lines.append(section["code"].rstrip())
                lines.append("```")
                lines.append("")

        filepath = os.path.join(full_dir, f"{topic_id}.md")
        with open(filepath, "w") as f:
            f.write("\n".join(lines))
        print(f"  Generated: {dir_name}/{topic_id}.md")

    print(f"\nGenerated {len(topics)} markdown files in {out_dir}/")

if __name__ == "__main__":
    main()
