use lazy_regex::regex;
use std::{
    collections::HashSet,
    env,
    ffi::OsStr,
    fs::{self, File},
    io::Read,
    path::{Path, PathBuf},
    process::Command,
};
use tree_sitter::{Parser, Point, Tree};

macro_rules! warn {
    ($message:expr) => {
        println!("::warning::{}", $message);
    };
    ($message:expr, $file:expr, $line:expr) => {
        println!("::warning file={},line={}::{}", $file, $line, $message);
    };
}

const KINDS: [&str; 7] = [
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
        .current_dir(path.parent().expect("path should have a parent"))
        .output()
        .expect("failed to execute git diff");
    let output = String::from_utf8(diff.stdout).expect("stdout should be utf-8");

    let regex = regex!(r"(?m)^@@ \-\d+(?:,\d+)* \+(\d+)(?:,\d+)* @@");
    let captures = regex.captures_iter(&output);
    return captures
        .map(|capture| {
            capture[1]
                .parse::<usize>()
                .expect("capture should be a line number")
        })
        .collect();
}

fn comments(tree: &Tree, point: Point) -> Vec<Point> {
    let mut cursor = tree.walk();
    let mut comments = Vec::new();
    loop {
        let node = cursor.node();
        if KINDS.contains(&node.kind()) {
            // TODO: Check for start/end of multiline comments.
            if let Some(comment) = cursor
                .node()
                .prev_named_sibling()
                .filter(|sibling| sibling.kind() == "comment")
            {
                comments.push(comment.start_position());
            }
        }
        cursor.goto_first_child_for_point(point);
        if cursor.node() == node {
            break;
        }
    }
    return comments;
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

    let comments: HashSet<Point> = lines
        .iter()
        .flat_map(|line| comments(&tree, Point::new(*line, 0)))
        .collect();

    let comments_lines: HashSet<usize> = comments.iter().map(|point| point.row).collect();
    for line in comments_lines.difference(&lines) {
        warn!("Possible stale comment", path.display(), line);
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
        .map(|arg| {
            Path::new(&arg)
                .canonicalize()
                .expect("arg should be a valid path")
                .to_owned()
        })
        .collect();
    for path in paths {
        find_recursively(&path);
    }
}
