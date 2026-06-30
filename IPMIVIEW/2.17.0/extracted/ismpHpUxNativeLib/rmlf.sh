#!/sbin/sh
# InstallShield (R)
# (c)2005, Macrovision Corporation
# (c)1996-2003, InstallShield Software Corporation
# (c)1990-1996, InstallShield Corporation
# All Rights Reserved

PATH=/usr/bin
PROGNAME=`basename $0`
LISTNAME=""
RECURSELISTNAME=""
typeset -i ATTEMPTS=10
typeset -i DEBUG=0
typeset -i SECONDS=2
typeset -i NFILES=0
typeset -i NDIRS=0
typeset -i SIZE=1000
typeset FILENAMES[0]=""
typeset DIRNAMES[0]=""

function Usage {
    echo "usage: $PROGNAME [-d] [-a <attempts>] [-t <wait_time>] -F <filelist> -R <dirlist>"
    exit 1
}

function DbgPrint {
    if [[ $DEBUG != 0 ]] then
        echo "$@"
    fi
}

# Process command line arguments
set -- `getopt F:R:da:t: $*`
if [[ $? -ne 0 ]] then
    Usage
fi
while [[ $# -gt 0 ]] do
    case $1 in
        (-F) LISTNAME=$2
             shift 2
             ;;
        (-R) RECURSELISTNAME=$2
             shift 2
             ;;
        (-a) ATTEMPTS=$2
             shift 2
             ;;
        (-d) DEBUG=1
             shift
             ;;
        (-t) SECONDS=$2
             shift 2
             ;;
        (--) shift
             break
             ;;
    esac
done
if [[ $DEBUG != 0 ]] then
    cp $LISTNAME /tmp/rmlf.in
    if [[ -f $RECURSELISTNAME ]] then
        cp $RECURSELISTNAME /tmp/rmlf_recurse.in
    fi
    exec 2>&1 1> /tmp/rmlf.out
fi
if [[ $LISTNAME = "" ]] then
    Usage
fi

DbgPrint "DEBUG    = $DEBUG"
DbgPrint "LISTNAME = $LISTNAME"
DbgPrint "RECURSELISTNAME = $RECURSELISTNAME"
DbgPrint "ATTEMPTS = $ATTEMPTS"
DbgPrint "SECONDS  = $SECONDS"

# Cleanup function is used to delete the files and directories. Configured number of(SIZE)files are deleted at a time.

function Cleanup {	
	let i=0	    
    while ((i<=NFILES)) ; do	    	
        NAME="${FILENAMES[i]}"	        
        if [[ "$NAME" != "" ]] then
            if [[ -d "$NAME" ]] then
                DbgPrint rmdir -f "$NAME"
                rmdir -f "$NAME"	                	                 
            else
                DbgPrint rm -f "$NAME"
                rm -f "$NAME"	                	                
            fi
            if [[ $? -eq 0 ]] then
                FILENAMES[i]=""
            else
                DbgPrint "Could not delete ${FILENAMES[i]}"
                STATUS=more
            fi
        fi
        let i=i+1
    done
}

sleep 4
cd /
while read NAME1 ; do
	if [ $NFILES -eq $SIZE ]
	then						
		Cleanup
		NFILES=0				
	fi		
    FILENAMES[NFILES]="$NAME1"
    let NFILES=NFILES+1
done < $LISTNAME
Cleanup


# Sort the list of dirs to be deleted recursively in reverse order so
# that ordinary files appear before the directory that contains them.
# Then read in the sorted list.
if [[ -f $RECURSELISTNAME ]] then
    sort -r $RECURSELISTNAME | \
    while read NAME ; do
        DIRNAMES[NDIRS]="$NAME"
        let NDIRS=NDIRS+1
    done
fi

# The ISMP uninstaller invokes this script while the current working
# directory is /<product_dir>/_uninst.  HP-UX will not allow us to remove
# the current working directory nor its parent directory.  So we change
# the current working directory to /.  Changing to / has no effect on
# the actual files and directories that get removed because they are all
# specified as absolute path names.
cd /

# Delete or attempt to delete the files.
typeset -i i
STATUS=more
while [[ $STATUS = more ]] do
    STATUS=none
    
    # Attempt to recursively delete directories. We do this first because
    # some of these directories are likely subdirectories of entries in the
    # FILENAMES list (which can contain directories that are not
    # recursively deleted). So, if we don't remove these first, the
    # directories in FILENAMES could likely not be uninstalled.
    let i=0
    while ((i<NDIRS)) ; do
        NAME="${DIRNAMES[i]}"
        if [[ "$NAME" != "" ]] then
            if [[ -d "$NAME" ]] then
                DbgPrint rm -rf "$NAME"
                rm -rf "$NAME"
            fi
            if [[ $? -eq 0 ]] then
                DIRNAMES[i]=""
            else
                DbgPrint "Could not delete directory ${DIRNAMES[i]}"
                STATUS=more
            fi
        fi
        let i=i+1
    done

    let i=0
    while ((i<NFILES)) ; do
        NAME="${FILENAMES[i]}"
        if [[ "$NAME" != "" ]] then
            if [[ -d "$NAME" ]] then
                DbgPrint rmdir -f "$NAME"
                rmdir -f "$NAME"
            else
                DbgPrint rm -f "$NAME"
                rm -f "$NAME"
            fi
            if [[ $? -eq 0 ]] then
                FILENAMES[i]=""
            else
                DbgPrint "Could not delete ${FILENAMES[i]}"
                STATUS=more
            fi
        fi
        let i=i+1
    done

    if [[ $STATUS = more ]] then
        # If we have tried to delete the files more than ATTEMPTS times,
        # set STATUS=quit.  Else sleep for a while and try to delete
        # them again.
        let ATTEMPTS=ATTEMPTS-1
        if [[ $ATTEMPTS -eq 0 ]] then
            STATUS=quit
        else
            DbgPrint "Waiting for $SECONDS seconds"
            sleep $SECONDS
        fi
    fi
done

# Put the name of each file and dir that couldn't be deleted into the SD cleanup
# file.  The cleanup file will be processed at the next system reboot.
if [[ $STATUS = quit ]] then
    umask 022
    let i=0
    while ((i<NFILES)) ; do
        NAME=${FILENAMES[i]}
        if [[ $NAME != "" ]] then
            echo $NAME >> /var/adm/sw/cleanupfile
        fi
        let i=i+1
    done

    let i=0
    while ((i<NDIRS)) ; do
        NAME=${DIRNAMES[i]}
        if [[ $NAME != "" ]] then
            echo $NAME >> /var/adm/sw/cleanupfile
        fi
        let i=i+1
    done
fi

let PARENT=`dirname "$0"`
if [[ -f "$0" ]] then 
	rm -f "$0"
fi 
let isTemp=`awk 'END { if (length(a) > 1 &&  index(a,"ismp") > 1) { print 1 } else { print 0} }' a="$PARENT" </dev/null  2>/dev/null`
if [[ "$isTemp" -eq 1 ]]  then  
		if [[ -d "$PARENT" ]]  then 
			rm -fr "$PARENT"
		fi 
fi  

exit 0
