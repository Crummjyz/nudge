use clap::{Arg, Command as App};
use lazy_regex::regex;
use std::{
    collections::HashSet,
    fs::File,
    io::Read,
    ops::Range,
    path::{Path, PathBuf},
    process::Command,
};
use tree_sitter::{Language, Node, Parser, Point, TreeCursor};
use walkdir::WalkDir;

macro_rules! warn {
    ($message:expr) => {
        println!("::warning::{}", $message);
    };
    ($message:expr, $file:expr, $lines:expr) => {
        println!(
            "::warning file={},line={},endline={}::{}",
            $file,
            $lines.start,
            $lines.end - 1,
            $message
        );
    };
}

trait Line {
    fn new_with_line(line: usize) -> Point;
    fn line(&self) -> usize;
}

impl Line for Point {
    fn new_with_line(line: usize) -> Point {
        Point::new(line - 1, 0)
    }

    fn line(&self) -> usize {
        self.row + 1
    }
}

fn diff(path: &Path, range: &String) -> HashSet<usize> {
    let path = &path.canonicalize().expect("path should exist");
    let diff = Command::new("git")
        .arg("diff")
        .arg("-U0")
        .arg(range)
        .arg(path)
        .current_dir(path.parent().expect("path should be in a git repo"))
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
                .unwrap_or(1)
                .max(1);
            start..(start + len)
        })
        .collect()
}

fn find_comments(cursor: &mut TreeCursor, point: Point) -> Vec<Range<usize>> {
    if cursor.goto_first_child_for_point(point).is_some() {
        fn leading_comment(node: Node) -> Node {
            node.prev_named_sibling()
                .filter(|sibling| sibling.kind() == "comment")
                .map_or(node, leading_comment)
        }

        let node = cursor.node();
        let comment = leading_comment(cursor.node());
        let mut comments = find_comments(cursor, point);
        if node.kind() != "comment" && comment != node {
            // TODO: it's inefficient to call leading_comment before the kind check.
            let start = comment.start_position().line();
            let end = node.start_position().line();
            comments.push(start..end);
        }
        comments
    } else {
        Vec::new()
    }
}

fn check_file(path: &Path, range: &String, language: Language) -> Result<(), std::io::Error> {
    let diff = diff(&path, &range);

    let mut source = String::new();
    File::open(path)?.read_to_string(&mut source)?;

    let mut parser = Parser::new();
    parser.set_language(language).unwrap();
    let tree = parser
        .parse(&source, None)
        .expect("source code should parse");

    let comments: HashSet<Range<usize>> = diff
        .iter()
        .flat_map(|line| find_comments(&mut tree.walk(), Point::new_with_line(*line)))
        .collect();

    for comment in comments {
        if !diff.iter().any(|line| comment.contains(line)) {
            warn!("Documentation may be stale", path.display(), comment);
        }
    }

    Ok(())
}

fn check_recursively(path: &Path, range: &String) {
    rayon::scope(|scope| {
        for entry in WalkDir::new(path) {
            let entry = entry.unwrap();
            let path = entry.path().to_owned();
            if let Some(language) = match path.extension().and_then(|ext| ext.to_str()) {
                Some("swift") => Some(tree_sitter_swift::language()),
                Some("rs") => Some(tree_sitter_rust::language()),
                Some("go") => Some(tree_sitter_go::language()),
                _ => None,
            } {
                scope.spawn(move |_| {
                    check_file(&path, range, language).unwrap_or_else(|err| {
                        eprintln!("Problem checking {}: {err}", path.display())
                    })
                });
            }
        }
    });
}

fn main() {
    let args = App::new("nudge")
        .about("Spot when implementations change, but docs don't.")
        .author("Finn Eger <finn@fantail.dev>")
        .version(env!("CARGO_PKG_VERSION"))
        .args([
            Arg::new("diff")
                .short('d')
                .help("Check a commit range")
                .value_name("RANGE")
                .default_value("HEAD~")
                .value_parser(clap::value_parser!(String)),
            Arg::new("path")
                .num_args(1..)
                .default_value(".")
                .value_parser(clap::value_parser!(PathBuf)),
        ])
        .get_matches();

    let range: &String = args.get_one("diff").expect("should default");
    let paths: Vec<&PathBuf> = args.get_many("path").expect("should default").collect();

    for path in paths {
        check_recursively(path, range);
    }
}
