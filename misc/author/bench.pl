#!/usr/bin/env perl
use 5.014;
use warnings;
use utf8;
use autodie;
use Time::HiRes qw/gettimeofday tv_interval/;

# you need `ghg` command to work it. ref. https://github.com/Songmu/ghg
my $workdir = '.tmp';
my $trial = 30;
my @vers = reverse qw/
    v0.17.4
    v0.17.3
    v0.14.2
    v0.14.1
    v0.12.8
    v0.12.6
    v0.9.0
    v0.8.0
/;

$ENV{GHG_HOME} = $workdir;
sub logf {
    my $msg = sprintf(shift, @_);
    $msg .= "\n" if $msg !~ /\n\z/ms;
    print STDERR $msg;
}

for my $ver (@vers) {
    my $bin = "$workdir/bin/ghq-$ver";
    next if -f $bin;

    logf 'installing ghq %s', $ver;
    my $cmd = "ghg get -u motemen/ghq\@$ver";
    `$cmd > /dev/null 2>&1`;
    rename "$workdir/bin/ghq", $bin;
}

my $VCS_GIT = '-vcs=git';
logf 'starting benchmark';
chomp(my $repos = `$workdir/bin/ghq-v0.14.2 list | wc -l`);
logf '%d repositories on local', $repos;
my %results;
for my $i (1..$trial) {
    logf 'trial: %d', $i;
    for my $ver (@vers) {
        my $bin = "$workdir/bin/ghq-$ver";
        for my $opt ('', $VCS_GIT) {
            if ($opt eq $VCS_GIT) {
                if ($ver !~ /^v0\.1/ || $ver lt 'v0.11.1') {
                    next;
                }
            }
            my $t = [gettimeofday];
            `$bin list $opt > /dev/null`;
            my $elapsed = tv_interval $t;
            if ($? > 0) {
                die 'error occurred on version: '. $ver;
            }
            $results{$ver . $opt} += $elapsed;
        }
    }
}

for my $ver (@vers) {
    my $val = $results{$ver};
    printf "%s: %0.5f\n", $ver, ($val/$trial);
    my $ver_with_git = $ver . $VCS_GIT;
    if (my $val = $results{$ver_with_git}) {
        printf "%s: %0.5f\n", $ver_with_git, ($val/$trial);
    }
}
