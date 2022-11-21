use regex::Regex;
use std::{collections::HashSet, env, fs::File, io::Read, path::Path, process::Command};
use tree_sitter::{Parser, Point, Tree};

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
        .current_dir(path.parent().unwrap())
        .output()
        .unwrap();
    let output = String::from_utf8(diff.stdout).unwrap();

    let regex = Regex::new(r"(?m)^@@ \-\d+(?:,\d+)* \+(\d+)(?:,\d+)* @@").unwrap();
    let captures = regex.captures_iter(&output);
    return captures
        .map(|capture| capture[1].parse::<usize>().unwrap())
        .collect();
}

fn comments(tree: &Tree, point: Point) -> Vec<Point> {
    let mut cursor = tree.walk();
    let mut comments = Vec::new();
    loop {
        let node = cursor.node();
        if KINDS.contains(&node.kind()) {
            // TODO: Check for start/end of multiline comments.
            let comment = cursor.node().prev_named_sibling().unwrap();
            if comment.kind() == "comment" {
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

fn main() {
    let path = Path::new(&env::args().nth(1).unwrap()).to_owned();
    let lines = diff(&path);

    let mut file = File::open(path).unwrap();
    let mut source = String::new();
    file.read_to_string(&mut source).unwrap();

    let mut parser = Parser::new();
    parser.set_language(tree_sitter_swift::language()).unwrap();
    let tree = parser.parse(&source, None).unwrap();

    let comments: HashSet<Point> = lines
        .iter()
        .flat_map(|line| comments(&tree, Point::new(*line, 0)))
        .collect();

    let comments_lines: HashSet<usize> = comments.iter().map(|point| point.row).collect();
    println!("{:#?}", comments_lines.difference(&lines));
}
