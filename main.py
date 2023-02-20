import matplotlib.pyplot as plt
import matplotlib
matplotlib.use('TkAgg')
import numpy as np
import pandas as pd
from matplotlib.ticker import MultipleLocator

#open csv file
def open_csv():
    data = np.genfromtxt('data.csv', delimiter=',')
    return data

def main():
    df = pd.read_csv('bench.csv', header=None)
    df.columns=["protocol","messages","size","time"]
    
    df["time"]=df["time"]/1000
    colors = {'quic':'red', 'tcp':'blue'}
   # colors=["red","blue"]
    #remove rows with 0 size and 0 time
    df=df[df["size"] != 0]
    df=df[df["size"] != 0]

    
  
    #plot the data x axis is the size of the message and y axis is the time, one line for
    #each protocol in diferent colors
    #interpolate dataframe dat

    fig, ax = plt.subplots()
    protocol_groups=df.groupby(['protocol'])
    labels=[]
    positions=[]
    offset=-10
    bps=[]
    for key, grp in protocol_groups:
        labels.append(key)
        grup=grp.groupby(['size'])
        val=np.array(list(grup.groups.keys()))
        
        positions.append(val)
        
        
        data= [np.array(grup.get_group(x)["time"]) for x in grup.groups]
        color=colors[key]
        bp=ax.boxplot(data,showfliers=False,positions=val+offset,widths=10,patch_artist=True)
        for element in ['boxes', 'whiskers', 'fliers', 'medians', 'caps']:
            plt.setp(bp[element], color=color)
        for patch in bp['boxes']:
            patch.set(facecolor="white")
        bps.append(bp)
        offset+=20
        
    #get longer list of positions
   
    #set legend for all bps
    ax.legend([bp["boxes"][0] for bp in bps],labels,loc="best")
    list_len = [len(i) for i in positions]
    positions= (positions[np.argmax(np.array(list_len))])
    ticks=np.arange(positions[0],positions[-1]+1,positions[1]-positions[0])
    
    ax.set_xticks(ticks)
    ax.set_xticklabels(ticks)
    
    

    #add legend and labels to each protocol
    
    ax.set_xlabel("Size (bytes)")
    ax.set_ylabel("Time (ms)")
    ax.set_title("QUIC and TCP performance differences")
    
    
  
    
    
    

    #print dataframe
     
    plt.show()
    exit()
    for key, grp in df.groupby(['protocol']):
        ax = grp.plot(ax=ax, kind='scatter', x='size', y='time', c=colors[key], label=key)

    plt.legend(loc='best')
    plt.xlabel("Messages")
    plt.ylabel("Time")
    plt.show()
if __name__ == "__main__":
    main()
