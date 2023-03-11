#[macro_export]
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

#[macro_export]
macro_rules! flag {
    ($flag:expr, $action:expr) => {
        if $flag.load(Ordering::Relaxed) {
            $action
        }
    };
}
