use lazy_regex::regex;
use std::{
    collections::HashSet,
    env,
    ffi::OsStr,
    fs::{self, File},
    io::Read,
    ops::Range,
    path::{Path, PathBuf},
    process::Command,
};
use tree_sitter::{Node, Parser, Point, Tree};

macro_rules! warn {
    ($message:expr) => {
        println!("::warning::{}", $message);
    };
    ($message:expr, $file:expr, $line:expr) => {
        println!("::warning file={},line={}::{}", $file, $line, $message);
    };
}

trait Line {
    fn new_with_line(line: usize) -> Point;
    fn line(&self) -> usize;
}

impl Line for Point {
    fn new_with_line(line: usize) -> Point {
        Point::new(line + 1, 0)
    }

    fn line(&self) -> usize {
        self.row + 1
    }
}

const DOCUMENTABLE_KINDS: [&str; 7] = [
    "variable_declaration",
    "function_declaration",
    "enum_declaration",
    "struct_declaration",
    "class_declaration",
    "protocol_declaration",
    "initializer_declaration",
];

fn diff(path: &Path) -> HashSet<usize> {
    let diff = Command::new("git")
        .arg("diff")
        .arg("@~..@")
        .arg("--unified=0")
        .arg(path)
        .current_dir(
            path.canonicalize()
                .unwrap()
                .parent()
                .expect("file should be in a git repo"),
        )
        .output()
        .expect("failed to execute git diff");
    let output = String::from_utf8(diff.stdout).expect("diff should be utf-8");

    let regex = regex!(r"(?m)^@@ \-\d+(?:,\d+)* \+(\d+)(?:,(\d+))* @@");
    let captures = regex.captures_iter(&output);
    captures
        .flat_map(|capture| {
            let start = capture
                .get(1)
                .unwrap()
                .as_str()
                .parse()
                .expect("should match a line number");
            let len = capture
                .get(2)
                .and_then(|m| m.as_str().parse().ok())
                .unwrap_or(1);
            start..(start + len)
        })
        .collect()
}

fn comments(tree: &Tree, line: usize) -> HashSet<Range<usize>> {
    let point = Point::new_with_line(line);

    let mut cursor = tree.walk();
    let mut comments = HashSet::new();
    loop {
        let node = cursor.node();
        if DOCUMENTABLE_KINDS.contains(&node.kind()) {
            fn prev_comment(node: Node) -> Node {
                match node
                    .prev_named_sibling()
                    .filter(|sibling| sibling.kind() == "comment")
                {
                    Some(comment) => prev_comment(comment),
                    None => node,
                }
            }

            let start = prev_comment(node);
            if start != node {
                comments.insert((start.start_position().line())..(node.start_position().line()));
            }
        }

        cursor.goto_first_child_for_point(point);
        if cursor.node() == node {
            break;
        }
    }
    comments
}

fn find(path: &Path) {
    let lines = diff(path);

    let mut file = File::open(path).expect("file should exist");
    let mut source = String::new();
    file.read_to_string(&mut source)
        .expect("file source should be readable");

    let mut parser = Parser::new();
    parser.set_language(tree_sitter_swift::language()).unwrap();
    let tree = parser
        .parse(&source, None)
        .expect("source code should parse");

    let comments: HashSet<Range<usize>> = lines
        .iter()
        .flat_map(|line| comments(&tree, *line))
        .collect();

    for comment in comments {
        if !lines.iter().any(|line| comment.contains(line)) {
            warn!("Documentation may be stale", path.display(), comment.start);
        }
    }
}

fn find_recursively(path: &Path) {
    if path.is_dir() {
        for entry in fs::read_dir(path).expect("path should be a readable dir") {
            let entry = entry.unwrap();
            let path = entry.path();
            find_recursively(&path);
        }
    } else {
        if path.extension() == Some(OsStr::new("swift")) {
            find(&path)
        }
    }
}

fn main() {
    let paths: Vec<PathBuf> = env::args()
        .skip(1)
        .map(|arg| Path::new(&arg).to_owned())
        .collect();
    for path in paths {
        find_recursively(&path);
    }
}
