use regex::Regex;
use std::{env, path::Path, process::Command};

fn diff(path: &Path) -> Vec<usize> {
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

fn main() {
    let path = Path::new(&env::args().nth(1).unwrap()).to_owned();
    let lines = diff(&path);
    println!("{lines:?}")
}
